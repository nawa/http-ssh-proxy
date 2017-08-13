package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config : proxy server config
type Config struct {
	AppPort   int             `yaml:"app-port"`
	StartPage string          `yaml:"start-page"`
	Hosts     map[string]Host `yaml:"hosts"`
}

// Host : host definition in config
type Host struct {
	Address    string      `yaml:"address"`
	Forwarding *Forwarding `yaml:"forwarding"`
}

// Forwarding : forwarding definition for host
type Forwarding struct {
	PrivateKey *string `yaml:"private-key"`
	Password   *string `yaml:"password"`
	User       string  `yaml:"user"`
	Server     string  `yaml:"server"`
}

// FromFile : creates config from file
func FromFile(configLocation string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return nil, fmt.Errorf("Can't read config file: %v", err)
	}
	cfg, err := NewConfig(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("Invalid config: %v", err)
	}
	return cfg, nil
}

// NewConfig : creates config from bytes
func NewConfig(yml []byte) (cfg *Config, err error) {
	cfg = new(Config)
	err = yaml.Unmarshal(yml, cfg)
	return
}
