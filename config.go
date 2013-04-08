package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	FeedbackTimeout int
	InterfaceKits   []*InterfaceKit
	Secret          string
}

func NewConfig() *Config {
	return &Config{FeedbackTimeout: 2000, InterfaceKits: make([]*InterfaceKit, 0)}
}

func LoadConfigFromFile(path string) (*Config, error) {
	var b []byte
	var err error

	c := NewConfig()

	if b, err = ioutil.ReadFile(path); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Doors() []*Door {
	var doors []*Door

	for _, k := range c.InterfaceKits {
		doors = append(doors, k.Doors...)
	}

	return doors
}

func (c *Config) SaveToFile(path string) error {
	var b []byte
	var err error

	if b, err = json.MarshalIndent(c, "", "  "); err != nil {
		return err
	}

	if err = ioutil.WriteFile(path, b, 0600); err != nil {
		return err
	}

	return nil
}
