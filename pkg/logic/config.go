package logic

import (
	"encoding/json"
	"fmt"
	"os"
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

type configRaw struct {
	Translations []struct {
		Regex   string `json:"regex"`
		Replace string `json:"replace"`
	} `json:"translations"`
}

// LoadConfig reads the configuration of the app from the current directory
func LoadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw := configRaw{}
	err = json.Unmarshal(content, &raw)
	if err != nil {
		return nil, err
	}

	if len(raw.Translations) < 1 {
		return nil, fmt.Errorf("%s: no translation found", path)
	}

	conf := Config{
		translations: make([]configEntry, len(raw.Translations)),
	}
	for i, entry := range raw.Translations {
		regex, err := regexp.Compile(entry.Regex)
		if err != nil {
			return nil, err
		}

		conf.translations[i] = configEntry{
			regex:   regex,
			replace: entry.Replace,
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
