package main

import (
	"github.com/greatoldcactus/textgenerationapi"
)

type Bot struct {
	Api *textgenerationapi.LLMApi
}

func NewBot(api *textgenerationapi.LLMApi) Bot {
	result := Bot{
		Api: api,
	}

	return result
}
