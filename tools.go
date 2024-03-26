package main

import (
	"fmt"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/greatoldcactus/textgenerationapi"
)

type GenerationData struct {
	Msg       textgenerationapi.Message
	User_data *UserData
}

var (
	GenerationChannel = make(chan GenerationData)
	MessageChan       = make(chan tgbotapi.MessageConfig)
)

func MessagingLoop(bot *tgbotapi.BotAPI) {
	for msg := range MessageChan {
		_, err := bot.Send(msg)
		if err != nil {
			err_text := fmt.Sprintf("error happened when trying to seng message: %s\n", err.Error())
			fmt.Println(err_text)
			msg := tgbotapi.NewMessage(msg.ChatID, err_text)
			bot.Send(msg)
		}
	}
}

func Generate(msg textgenerationapi.Message, user_data *UserData) (result textgenerationapi.Message, err error) {

	switch api_typed := api.(type) {
	case textgenerationapi.LLMApiModels:
		api_typed.SetModel(user_data.Model)
	}
	switch user_data.Mode {
	case ModeChat:
		user_data.History.Add(msg)
		api.SetAnswerTokens(int(user_data.MaxTokens))
		msg, err = api.Generate(user_data.History)
		if err != nil {
			err = fmt.Errorf("failed to chat: %w", err)
			return
		}
	case ModeContinue:
		result, err = api.Continue(msg)
		if err != nil {
			err = fmt.Errorf("failed to continue: %w", err)
			return
		}
		user_data.History.Clear()
	case ModeSingleMessage:
		result, err = api.Answer(msg)
		if err != nil {
			err = fmt.Errorf("failed to answer: %w", err)
			return
		}
		user_data.History.Clear()
	default:
		err = ErrUnknownMode
		return
	}

	user_data.History.Add(msg)
	user_data.History.Add(result)

	return
}

func GeneratingLoop(bot *tgbotapi.BotAPI) {
	for gen_data := range GenerationChannel {
		result, err := Generate(gen_data.Msg, gen_data.User_data)
		gen_data.User_data.StoreSimple()
		if err != nil {
			msg := tgbotapi.NewMessage(gen_data.User_data.UserID, fmt.Sprintf("error happened on generation %s", err.Error()))
			MessageChan <- msg
		} else {
			msg := tgbotapi.NewMessage(gen_data.User_data.UserID, result.Message)
			MessageChan <- msg
		}

	}
}

func CreateSelectKeyboard(bot *tgbotapi.BotAPI, userdata *UserData) {
	switch api.(type) {
	case textgenerationapi.LLMApiModels:
	default:
		return
	}
	api_models := api.(textgenerationapi.LLMApiModels)
	models, err := api_models.ListModels()
	if err != nil {
		msg := tgbotapi.NewMessage(userdata.UserID, fmt.Sprintf("failed to list models: %v", err))
		MessageChan <- msg
		return
	}

	msg := tgbotapi.NewMessage(userdata.UserID, "choose model: ")

	//TODO pagination?
	buttons := [][]tgbotapi.InlineKeyboardButton{}
	for _, model := range models {
		data_string := fmt.Sprintf("model!%s", model)
		button := tgbotapi.NewInlineKeyboardButtonData(model, data_string)
		row := tgbotapi.NewInlineKeyboardRow(button)

		buttons = append(buttons, row)

	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg.ReplyMarkup = keyboard

	MessageChan <- msg

}
