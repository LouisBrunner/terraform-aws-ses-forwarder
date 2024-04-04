package logic

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type configEntry struct {
	regex   *regexp.Regexp
	replace string
}

// Config contains the configuration of the app
type Config struct {
	translations []configEntry
}

type configEmail struct {
	Regex     string   `json:"regex"`
	ForwardTo []string `json:"forward_to"`
}

type config struct {
	Emails []configEmail `json:"emails"`
}

// LoadConfig reads the configuration of the app from the current directory
func LoadConfig(content string) (*Config, error) {
	raw := config{}
	err := json.Unmarshal([]byte(content), &raw)
	if err != nil {
		return nil, err
	}

	if len(raw.Emails) < 1 {
		return nil, fmt.Errorf("no translation found")
	}

	conf := Config{
		translations: make([]configEntry, 0, len(raw.Emails)),
	}
	for _, entry := range raw.Emails {
		regex, err := regexp.Compile(entry.Regex)
		if err != nil {
			return nil, err
		}

		for _, alias := range entry.ForwardTo {
			conf.translations = append(conf.translations, configEntry{
				regex:   regex,
				replace: alias,
			})
		}
	}

	return &conf, nil
}

// Map will map the given recipient to a new one (or return an error otherwise)
func (conf *Config) Map(to string) (string, error) {
	for _, entry := range conf.translations {
		if entry.regex.MatchString(to) {
			return entry.regex.ReplaceAllString(to, entry.replace), nil
		}
	}
	return "", fmt.Errorf("no match found for `%s`", to)
}
