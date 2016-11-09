package config

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestLoad(t *testing.T) {
	testData := Config{
		Port:          ":8000",
		Timeout:       0,
		Debug:         false,
		AllowedOrigin: "*",
	}
	ymlString, _ := yaml.Marshal(testData)
	_, err := Load(bytes.NewReader(ymlString))
	if err != nil {
		t.Errorf("Load mustn't return an error if loaded config file is right, got '%s'", err.Error())
	}
}

func TestLoadNoPort(t *testing.T) {
	testData := Config{
		Timeout:       0,
		Debug:         false,
		AllowedOrigin: "*",
	}
	ymlString, _ := yaml.Marshal(testData)
	_, err := Load(bytes.NewReader(ymlString))
	if err == nil {
		t.Errorf("Load must return an error if loaded config file has no port information")
	}
}

func TestLoadNoAllowedOrigin(t *testing.T) {
	testData := Config{
		Port:    ":8000",
		Timeout: 0,
		Debug:   false,
	}
	ymlString, _ := yaml.Marshal(testData)
	_, err := Load(bytes.NewReader(ymlString))
	if err == nil {
		t.Errorf("Load must return an error if loaded config file has no allowed origin information")
	}
}
