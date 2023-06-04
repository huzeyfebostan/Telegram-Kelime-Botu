package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/huzeyfebostan/go-telegram-bot/database"
	"github.com/huzeyfebostan/go-telegram-bot/handlers"
	"log"
	"os"
)

func main() {

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Telegram bot token'ı ortam değişkenlerinde bulunamadı.")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("%s'u başarıyla başlatıldı!", bot.Self.UserName)

	database.ConnectDB()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 120

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "👋 Merhaba "+update.Message.From.FirstName+", Ben Kelime Botu\n\n Lütfen Yapmak istediğiniz işlemi seçin;")
				msg.ReplyMarkup = handlers.CreateMainMenu()
				bot.Send(msg)
			} else {
				handlers.HandleMessage(bot, update.Message)
			}
		} else if update.CallbackQuery != nil {
			handlers.HandleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}
