package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)


func main() {
	// Set up logging
    log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	// Prompt for yml file, if any
	fmt.Print("Config file name (press enter to skip): ")
	var configFileName string
	fmt.Scanln(&configFileName)

	// Read the configuration file
	config, missing, err := parseConfig(configFileName)
	if err != nil {
		log.Error().Err(err).Msg("Config filename supplied missing or incorrect")
		os.Exit(1)
	}

	// Request inputs for any config items not supplied
	for _, item :=range missing {
		var input string
		fmt.Printf("%s: ", item)
		fmt.Scanln(&input)
		switch item {
		case "Subdomain":
			config.subdomain = input
		case "Username":
			config.username = input
		case "Password":
			config.password = input
		case "Translations filename":
			config.translationsFile = input
		}
	}

	// Get Locale-Ids for instance
	locales, err := getLocales(config)
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to Zendesk instance or parse installed languages. Check subdomain is correct")
		os.Exit(1)
	}

	// Read the csv file containing translations
	translations, err := parseTranslations(config.translationsFile, locales)
	if err != nil {
		log.Error().Err(err).Msg("Translations CSV file missing or incorrect")
		os.Exit(1)
	}

	// Loop through translations
	for _, translation := range translations {
		// Get correctly formatted payload
		translationPayload, err := translation.dynamicContentPayload()
		if err != nil {
			log.Error().Err(err).Msg("Error parsing translations")
			os.Exit(2)
		}

		// Send payload to Zendesk
		err = postToZendesk(translationPayload, config)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Error uploading translation to Zendesk: %s", translation.Name))
		}
	}
}