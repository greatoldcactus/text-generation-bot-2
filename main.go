package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/greatoldcactus/textgenerationapi"
)

func main() {
	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	go MessagingLoop(bot)
	go GeneratingLoop(bot)

	// в канал updates прилетают структуры типа Update
	// вычитываем их и обрабатываем
	for update := range updates {
		// универсальный ответ на любое сообщение
		reply := "Не знаю что сказать"

		if update.Message == nil {

			if update.CallbackQuery == nil {
				continue
			}
			go func() {
				data := update.CallbackQuery.Data
				callback_parts := strings.Split(data, "!")
				if len(callback_parts) < 1 {
					fmt.Println("error happened when trying to process callback: unable to detect type")
					return
				}
				callback_type := callback_parts[0]

				callback_payload := strings.TrimPrefix(data, fmt.Sprintf("%s!", callback_type))

				switch callback_type {
				case "model":
					user, err := GetUser(update.CallbackQuery.From.UserName, int64(update.CallbackQuery.From.ID))
					if err != nil {
						fmt.Println("failed to get user to set model ", err)
						return
					}
					user.Model = callback_payload
					user.StoreSimple()
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("model selected: %s", callback_payload))
					MessageChan <- msg

				}
			}()
			continue

		}

		chat := update.Message.Chat
		_, err := GetUser(chat.UserName, chat.ID)
		if err != nil {
			_, err := NewUser(chat.UserName, chat.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "failed to create new user")
				MessageChan <- msg
				continue
			}
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			reply = "hello there!"

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			MessageChan <- msg
		case "help":
			reply = `
/start - start dialog
/help  - show this message
/list  - list models
/model - select model
/profile - get user data
/clear - clear history
/tokens - sets token cnt
			`

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			MessageChan <- msg
		case "list":
			go func() {
				switch api_typed := api.(type) {
				case textgenerationapi.LLMApiModels:
					models, err := api_typed.ListModels()
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("failed to list models: %s", err.Error()))
						MessageChan <- msg
						return
					}
					reply := strings.Join(models, "\n")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					MessageChan <- msg
					return
				}
				reply := "model selection not implemented"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
				MessageChan <- msg
			}()
		case "model":
			user, err := GetUser(chat.UserName, chat.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("failed to find user: %s", err.Error()))
				MessageChan <- msg
				continue
			}
			go CreateSelectKeyboard(bot, user)
			user.StoreSimple()
		case "tokens":
			user, err := GetUser(update.Message.Chat.UserName, update.Message.Chat.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("unable to get user data: %s", err.Error()))
				MessageChan <- msg
				continue
			}
			new_token_cnt := update.Message.CommandArguments()
			if new_token_cnt == "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "add name after command")
				MessageChan <- msg
				continue
			}
			cnt, err := strconv.Atoi(new_token_cnt)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("failed to parse size: '%s'", new_token_cnt))
				MessageChan <- msg
				continue
			}
			user.MaxTokens = uint(cnt)
			user.StoreSimple()
		case "profile":
			user, err := GetUser(update.Message.Chat.UserName, update.Message.Chat.ID)
			fmt.Println(user.History.Messages)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("unable to get user data: %s", err.Error()))
				MessageChan <- msg
				continue
			}
			user_data, err := json.Marshal(user)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("unable to serialize user data: %s", err.Error()))
				MessageChan <- msg
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("user data: \n%s", user_data))
			MessageChan <- msg
		case "clear":
			user, err := GetUser(update.Message.Chat.UserName, update.Message.Chat.ID)
			fmt.Println(user.History.Messages)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("unable to get user dada for clear: %s", err.Error()))
				MessageChan <- msg
				continue
			}
			user.History.Clear()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "history cleared")
			MessageChan <- msg
			user.StoreSimple()

		case "":
			go func() {
				user, err := GetUser(update.Message.Chat.UserName, update.Message.Chat.ID)
				fmt.Println(user.History.Messages)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("unable to get user data: %s", err.Error()))
					MessageChan <- msg
					return
				}

				msg := textgenerationapi.Message{
					AuthorName: "user",
					Message:    update.Message.Text,
				}

				GenerationChannel <- GenerationData{
					Msg:       msg,
					User_data: user,
				}
			}()

		default:
			reply = fmt.Sprintf("unknown command: %s", update.Message.Command())
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			MessageChan <- msg
		}
	}
}
