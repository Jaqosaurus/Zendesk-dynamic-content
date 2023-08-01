# Dynamic content creator for Zendesk

## File format

### Config file

An optional yml file containing the subdomain, username and filename of the translations. You can also include a password or API key (both under the 'password' field). Any fields not included will be prompted for at runtime. An example config.yml file is included in this repo.

If using an API key, please append the username with /token. For example  
 `username: joe.bloggs@example.com/token`

subdomain should be just the subdomain part, if your instance is `https://example.zendesk.com` then you would put the following in the config file  
`subdomain: example`

There are 4 optional fields within this file, as shown below (with examples of how they should be filled)

```yml
subdomain: example
username: joe.bloggs@example.com/token
password: APIKEY
translationsFile: translations.csv
```

### CSV file with translations

CSV file with headings - heading row is 'name' and translations to follow. The second column in the document will be assumed to be the default language.  
Example (header row if default language is UK English and translations in French and Spanish are to be added): `name,en-gb,fr,es-es`

### Locales

Locales must be listed using their locale codes, such as 'en-gb' (English UK), 'fr' (French), 'es-es' (Spanish Spanish) - the full list is given [here](https://support.zendesk.com/hc/en-us/articles/4408821324826-Zendesk-language-support-by-product)

## Troubleshooting

Common troubleshooting advice

|Error | Things to try|
--- | ---
|"Config filename submitted missing or incorrect" | Check the location and name of the config file - it should be in the main folder and you should include the .yml extension (eg enter 'config,yml' not 'config'|
| "Unable to connect to Zendesk instance or parse installed languages" | The most likely reason behind this is the subdomain is invalid - please check you are only including the part before 'zendesk.com'. |
| "Translation CSV file missing or incorrect" | This could be that the format or encoding is incorrect. Check it is a CSV file. Some programs encode CSVs weirdly, if you aren't sure how to change the encoding copy it into google sheets and save a CSV from there as that always seems to work. |
| "Error uploading translations to Zendesk: [translation name]" | There should be a secondary error that will give more information where it's gone wrong, some examples below. Some are processed to shorter errors, some will dump the entire response from Zendesk and you'll have to read through the json to find out what's gone wrong. |
| Authentication errors, eg '401 - Couldn't authenticate you' | Check that the Zendesk instance is correctly configured to allow connections to the API in Admin Centre > Apps and Integrations > Zendesk API. There are individual toggles for Password and Token access to the API. If this is all set up the problem might be from the username and password file - If using a token please ensure you inlude '/token' at the end of your username. If using a config file, check for trailing spaces. You may also get this error if the subdemain is incorrect |
| Server errors | Unless there is a problem with Zendesk itself, this is likely a result of an improperly configured translations file that managed to not get caught earlier |
| "Locale not installed in Zendesk Instance" | Either the locale is not installed in the instance or you have not used the correct locale code - eg 'en-gb' instead of 'en-us' - check those which are installed in the Zendesk instance. Use codes only, not names in the CSV file and avoid trailing spaces.|
| "Name has already been taken"/"Name already used in dynamic content placeholder" | A dynamic content item with the same name alread exists, or the name is used in a dynamic content placeholder. If this isn't a duplication rename the version in your CSV or delete the entry in Zendesk if you wish to overwrite.|
| Other | Check there are no blank translation cells (no handling has been written, errors are expected). Check csv file only has one tab if you've written it in excel.

## Future plans

- Add handling for updating rather than just creating new translations
- Add handling for blank content values
- Package up nicely for others
