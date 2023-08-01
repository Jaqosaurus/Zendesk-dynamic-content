package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)


type config struct {
	subdomain string
	username string
	password string
	translationsFile string
}

type translation struct {
	Name string `json:"name"`
	DefaultLanguageId int `json:"locale_id"`
	Variants []variant `json:"variants"`
}

type variant struct {
	language string
	LanguageId int `json:"locale_id"`
	Default bool `default:"false" json:"default"`
	Content string `json:"content"`
}

type locale struct {
	Locale string
	Id int
}

type zendeskLocalesWrapper struct {
	Locales []locale
}

type zendeskDynamicContentWrapper struct {
	Item translation `json:"item"`
}

type zendeskErrorResponse struct {
	Error string `json:"error"`
	Description string `json:"description"`
	Details []string `json:"details"`
}

// Takes config file specified in parameter and returns a config stuct
func parseConfig(filename string) (config, []string, error)  {

	// Setup config map
	configMap := make(map[interface{}]interface{})

	// If a config file was supplied
	if filename != "" {
			// Read the configuration file
	configFile, err := ioutil.ReadFile(filename)
    if err != nil {
		return config{}, nil, err
    }
	
	// Unmarshal into the map
	err = yaml.Unmarshal(configFile,&configMap)
	if err != nil {
    	return config{}, nil, err
    }

	}

	// Load into Config struct
	var config config
	var found bool
	var missing []string

	config.subdomain, found = configMap["subdomain"].(string)
	if !found {
		missing = append(missing, "Subdomain")
	}
	config.username, found = configMap["username"].(string)
	if !found {
		missing = append(missing, "Username")
	}
	config.password, found = configMap["password"].(string)
	if !found {
		missing = append(missing, "Password")
	}
	config.translationsFile, found = configMap["translationsFile"].(string)
	if !found {
		missing = append(missing, "Translations filename")
	}

	return config, missing, nil
}

// takes the filename of a csv file containing translations and returns an array of translations
func parseTranslations(filename string, locales []locale) ([]translation, error) {

	// Open and read the CSV file
	translationsFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(translationsFile)
	
	// Loop through file contents and populate array of translations
	i := 0
	var translations []translation
	var headers []string

	for {
		//  Read the row
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Parse row into a translation, with variants
		var translation translation
		if i == 0 {
			for j :=0; j < len(row); j++ {
				headers = append(headers, row[j])
			}
		} else {
			for j :=0; j < len(row); j++ {
				if headers[j] == "name" {
					translation.Name = row[j]
				} else {
					var variant variant
					variant.language = headers[j]
					for i := range locales {
						if locales[i].Locale == headers[j] {
							variant.LanguageId = locales[i].Id
							if j == 1 {
								translation.DefaultLanguageId = locales[i].Id
							}
						}
					}
					variant.Content = row[j]
					if j == 1 {
						variant.Default = true
					}
					translation.Variants = append(translation.Variants, variant)
				}
				if j == 1 {

				}
			}
			translations = append(translations, translation)
		}
		i++
	}

	return translations, nil
}

// Function to get the Locales, and their ID's, which are active in the instance
func getLocales(config config) ([]locale, error) {

	// Connect to Zendesk. No auth required for this request
	url := "https://" + config.subdomain + ".zendesk.com/api/v2/locales"
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client {}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	jsonLocales := string(body)

	var data zendeskLocalesWrapper

	err = json.Unmarshal([]byte(jsonLocales), &data)
	if err != nil {
		return nil, err
	}

	return data.Locales, nil
}

// Method upon a translation struct, returns payload formatted for creating a new
// Dynamic Content payload in Zendesk
func (translation translation) dynamicContentPayload() (string, error) {

	translationWrapped := zendeskDynamicContentWrapper{Item: translation}

	translationJson, err := json.Marshal(translationWrapped)
	if err != nil {
		return "error", err
	}

	return string(translationJson), nil
}

// Send payload to Zendesk
func postToZendesk(payload string, config config) ( error) {

	url := "https://" + config.subdomain + ".zendesk.com/api/v2/dynamic_content/items"
	method := "POST"
	data := bytes.NewReader([]byte(payload))
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(config.username+":"+config.password)))

	client := &http.Client {}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		var errorResponse zendeskErrorResponse
		var errorMessage string
		err = json.Unmarshal([]byte(string(body)),&errorResponse)
		if res.StatusCode == 401 {
			errorMessage = errorResponse.Error
		} else if res.StatusCode == 404 {
			errorMessage = fmt.Sprintf("Subdomain %s.zendesk.com does not exist", config.subdomain)
		} else if strings.Contains(string(body),"Title: has already been taken") {
			errorMessage = "Name has already been taken"
		} else if strings.Contains(string(body),"}} is already in use"){
			errorMessage = "Name already used in dynamic content placeholder"
		} else if strings.Contains(string(body), "Translation locale invalid locale") {
			errorMessage = "One or more locales not installed in Zendesk Instance"
		} else {
			errorMessage = fmt.Sprintf("%+v", string(body))
		}
		return errors.New(fmt.Sprintf("%s - %s", res.Status, errorMessage))
	}

	return nil
}