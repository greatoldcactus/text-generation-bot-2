package main

import (
	"flag"
	"io"
	"log"
	"os"
)

var telegramBotToken string

func init() {
	// принимаем на входе флаг -telegrambottoken
	flag.StringVar(&telegramBotToken, "telegrambottoken", "", "Telegram Bot Token")
	flag.Parse()

	// без него не запускаемся
	if telegramBotToken == "" {
		file, err := os.Open("telegram.token")
		if err == nil {
			token, err := io.ReadAll(file)
			if err == nil {
				telegramBotToken = string(token)
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		panic("telegram token undefined")
	}
}
