package main

import (
	"encoding/json"
	"os"
)

var appConfig Config

type Config struct {
	Token string
}

func readConfig() {
	file, err := os.Open("config.json")

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)
	appConfig = *(new(Config))
	err = decoder.Decode(&appConfig)
	if err != nil {
		panic(err)
	}
}
