package app

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v1"
)

type Config struct {
	AWS AWS `yaml:"aws"`
}
type AWS struct {
	AccessKey     string `yaml:"access_key"`
	SecretKey     string `yaml:"secret_key"`
	AssumeRoleARN string `yaml:"assume_role_arn"`
	Region        string `yaml:"region"`
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
