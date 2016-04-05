package config

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestLoad(t *testing.T) {
	testData := struct {
		Port int
	}{
		Port: 8000,
	}
	ymlString, _ := yaml.Marshal(testData)
	_, err := Load(bytes.NewReader(ymlString))
	if err != nil {
		t.Errorf("Load mustn't return an error if loaded config file is right, got '%s'", err.Error())
	}
}

func TestLoadNoPort(t *testing.T) {
	testData := struct{}{}
	ymlString, _ := yaml.Marshal(testData)
	_, err := Load(bytes.NewReader(ymlString))
	if err == nil {
		t.Errorf("Load must return an error if loaded config file has no port information")
	}
}
