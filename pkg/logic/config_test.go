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

func TestConfigMap(t *testing.T) {
	conf := setupConfig()
	for _, testcase := range []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "works 1",
			input:    "abc@example.com",
			expected: "123",
		},
		{
			name:     "works 2",
			input:    "abc@def.ghi",
			expected: "def.abc",
		},
		{
			name:    "fails (not found)",
			input:   "@example.com",
			wantErr: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			to, err := conf.Map(testcase.input)
			if testcase.wantErr {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, testcase.expected, to)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		config      string
		mapInput    string
		mapExpected string
		wantErr     bool
	}{
		{
			name: "works",
			config: `{
	"emails": {".*@example.com":["123"]}
}`,
			mapInput:    "abc@example.com",
			mapExpected: "123",
		},
		{
			name:    "fails (invalid json)",
			config:  "123",
			wantErr: true,
		},
		{
			name: "fails (empty)",
			config: `{
  "emails": {".*@example.com":["123"]},
}`,
			wantErr: true,
		},
		{
			name: "fails (empty)",
			config: `{
  "emails": {}
}`,
			wantErr: true,
		},
		{
			name: "fails (invalid regex)",
			config: `{
  "emails": {"[":["123"]}
}`,
			wantErr: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			config, err := LoadConfig(testcase.config)
			if testcase.wantErr {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					to, err := config.Map(testcase.mapInput)
					if assert.NoError(t, err) {
						assert.Equal(t, testcase.mapExpected, to)
					}
				}
			}
		})
	}
}
