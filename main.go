package main

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Twitter TwitterConfig `yaml:"Twitter"`
}

var config Config

func readConfig() error {
	data, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		log.Error().Str("Init:", err.Error())
		return err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Config loaded from ./config.yml")
	return nil
}

func main() {
	log.Info().Msg("TetraFour Starting.")
	readConfig()
	config.Twitter.connect()
	log.Info().Msg("TetraFour Stopping.")
}
