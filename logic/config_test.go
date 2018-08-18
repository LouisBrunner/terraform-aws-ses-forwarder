package logic

import (
	"io/ioutil"
	"os"
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

func setupFile(content string) (string, func()) {
	file, _ := ioutil.TempFile("", "ef")
	file.WriteString(content)
	filename := file.Name()
	file.Close()
	return filename, func() {
		os.Remove(filename)
	}
}

func TestLoadConfig_Works(t *testing.T) {
	file, done := setupFile(`{
  "translations": [
    {"regex":".*@example.com","replace":"123"}
  ]
}`)
	defer done()
	conf, err := LoadConfig(file)
	if assert.NoError(t, err) {
		to, err := conf.Map("abc@example.com")
		if assert.NoError(t, err) {
			assert.Equal(t, "123", to)
		}
	}
}

func TestLoadConfig_Fails_NoFile(t *testing.T) {
	_, err := LoadConfig("123")
	assert.EqualError(t, err, "open 123: no such file or directory")
}

func TestLoadConfig_Fails_Invalid(t *testing.T) {
	file, done := setupFile(`{
  "translations": [
    {"regex":".*@example.com","replace":"123"},
  ]
}`)
	defer done()
	_, err := LoadConfig(file)
	assert.EqualError(t, err, "invalid character ']' looking for beginning of value")
}

func TestLoadConfig_Fails_Empty(t *testing.T) {
	file, done := setupFile(`{
  "translations": [
  ]
}`)
	defer done()
	_, err := LoadConfig(file)
	assert.EqualError(t, err, file+": no translation found")
}

func TestLoadConfig_Fails_InvalidRegex(t *testing.T) {
	file, done := setupFile(`{
  "translations": [
    {"regex":"[","replace":"123"}
  ]
}`)
	defer done()
	_, err := LoadConfig(file)
	assert.EqualError(t, err, "error parsing regexp: missing closing ]: `[`")
}
