package main

import (
	"encoding/json"
	"io"
	"os"

	textgenerationollamaapi "github.com/greatoldcactus/text-generation-ollama-api"
	"github.com/greatoldcactus/textgenerationapi"
)

type Config struct {
	Url string `json:"url"`
}

var config Config

func init() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	string_config, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(string_config, &config)
	if err != nil {
		panic(err)
	}
}

var api textgenerationapi.LLMApi

func init() {
	api = &textgenerationollamaapi.TextGenerationAPIOllama{
		Url: config.Url,
	}
}
