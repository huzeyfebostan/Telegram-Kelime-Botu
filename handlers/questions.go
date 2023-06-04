package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/huzeyfebostan/go-telegram-bot/database"
	"github.com/huzeyfebostan/go-telegram-bot/models"
	"gorm.io/gorm"
	"math/rand"
)

var Words = make(map[int64][]models.Word)
var correctAnswers = make(map[int64]string)
var userQuestions = make(map[int64][]Question)

type Question struct {
	Text    string
	Options []string
	Answer  string
}

func NextPage(userID int64) {
	UserPages[userID] = UserPages[userID] + 1
}

func PrevPage(userID int64) {
	if UserPages[userID] > 0 {
		UserPages[userID] = UserPages[userID] - 1
	}
}

func createQuestion(word models.Word, level string) Question {
	var incorrectWords []models.Word
	db := database.DB()
	db.Where("level = ? AND english != ?", level, word.English).Order(gorm.Expr("random()")).Limit(3).Find(&incorrectWords)

	question := fmt.Sprintf("'%s' kelimesinin Türkçe karşılığı nedir ?", word.English)
	options := []string{word.Turkish}
	for _, incorrectWord := range incorrectWords {
		options = append(options, incorrectWord.Turkish)
	}

	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	return Question{
		Text:    question,
		Options: options,
		Answer:  word.Turkish,
	}
}

func sendQuestion(bot *tgbotapi.BotAPI, chatID int64, question Question) {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, option := range question.Options {
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", option), "option_"+option)
		row := []tgbotapi.InlineKeyboardButton{button}
		rows = append(rows, row)
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, question.Text)
	msg.ReplyMarkup = &markup
	bot.Send(msg)

	correctAnswers[chatID] = question.Answer
}

func handleAnswer(bot *tgbotapi.BotAPI, chatID int64, selectedOption string) {
	correctAnswer := correctAnswers[chatID]
	if selectedOption == correctAnswer {
		msg := tgbotapi.NewMessage(chatID, "Doğru!")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Yanlış. Doğru cevap %s.", correctAnswer))
		bot.Send(msg)
	}
}
