package main

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
)

var config Config

type Config struct {
	Redis RedisConfig `yaml:"redis"`
	MinimalKeyLength int `yaml:"minimal-key-length"`
	ExpectedStripesPerHour int `yaml:"expected-stripes-per-hour"`
	AllowedBadAttempts int `yaml:"allowed-bad-attempts"`
	AppropriateChanceToGuess int `yaml:"appropriate-chance-to-guess"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func loadConfig() {
	config = Config{}
	addFileDataToConfig("default")
	addFileDataToConfig("config")
}

func addFileDataToConfig(name string) {
	dat, err := ioutil.ReadFile("config/"  + name + ".yaml")
	if err != nil {
		log.Fatal("Can't load config file config/"  + name + ".yaml: " + err.Error())
	}
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		log.Fatal("Can't parse config file config/"  + name + ".yaml: " + err.Error())
	}
}