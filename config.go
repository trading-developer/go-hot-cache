package main

import (
	"encoding/json"
	"os"
)

var appConfig Config

type Config struct {
	Token  string `json:"token"`
	ChatId string `json:"chatId"`
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

func setLogParam() {
	log.Out = os.Stdout

	file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Ошибка создания файла логов, используйте stderr")
	}
}
