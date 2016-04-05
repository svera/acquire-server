package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

// Config holds data needed to run a TBG server instance
type Config struct {
	Port string
}

// Load reads configuration from config.yml and parses it
func Load(src io.Reader) (*Config, error) {
	c := &Config{}
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return c, err
	}

	if err = yaml.Unmarshal(data, c); err != nil {
		return c, err
	}
	err = c.validate()
	return c, err
}

func (c *Config) validate() error {
	if c.Port == "" {
		return errors.New("TBG configuration: Invalid port")
	}
	return nil
}
