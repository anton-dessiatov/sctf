package app

import (
	"io/ioutil"
	"log"

	"github.com/anton-dessiatov/sctf/tf/terra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Terra Terra `yaml:"terra"`
}

type Terra struct {
	Credentials terra.Credentials `yaml:"credentials"`
}

func readConfig() Config {
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("ioutil.ReadFile: %w", err)
	}

	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("yaml.Unmarshal: %w", err)
	}

	return c
}
