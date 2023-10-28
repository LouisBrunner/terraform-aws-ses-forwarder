package logic

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupConfig() *Config {
	return &Config{
		translations: []configEntry{
			{regex: regexp.MustCompile(".+@example.com"), replace: "123"},
			{regex: regexp.MustCompile("abc@(def).ghi"), replace: "$1.abc"},
		},
	}
}

func TestConfigMap_Works_1(t *testing.T) {
	conf := setupConfig()
	to, err := conf.Map("abc@example.com")
	if assert.NoError(t, err) {
		assert.Equal(t, "123", to)
	}
}

func TestConfigMap_Works_2(t *testing.T) {
	conf := setupConfig()
	to, err := conf.Map("abc@def.ghi")
	if assert.NoError(t, err) {
		assert.Equal(t, "def.abc", to)
	}
}

func TestConfigMap_Fails_NotFound(t *testing.T) {
	conf := setupConfig()
	_, err := conf.Map("@example.com")
	assert.EqualError(t, err, "no match found for `@example.com`")
}

func TestLoadConfig_Works(t *testing.T) {
	conf, err := LoadConfig(`{
  "emails": {".*@example.com":["123"]}
}`)
	if assert.NoError(t, err) {
		to, err := conf.Map("abc@example.com")
		if assert.NoError(t, err) {
			assert.Equal(t, "123", to)
		}
	}
}

func TestLoadConfig_Fails_NoFile(t *testing.T) {
	_, err := LoadConfig("123")
	assert.EqualError(t, err, "json: cannot unmarshal number into Go value of type logic.config")
}

func TestLoadConfig_Fails_Invalid(t *testing.T) {
	_, err := LoadConfig(`{
  "emails": {".*@example.com":["123"]},
}`)
	assert.EqualError(t, err, "invalid character '}' looking for beginning of object key string")
}

func TestLoadConfig_Fails_Empty(t *testing.T) {
	_, err := LoadConfig(`{
  "emails": {}
}`)
	assert.EqualError(t, err, "no translation found")
}

func TestLoadConfig_Fails_InvalidRegex(t *testing.T) {
	_, err := LoadConfig(`{
  "emails": {"[":["123"]}
}`)
	assert.EqualError(t, err, "error parsing regexp: missing closing ]: `[`")
}
