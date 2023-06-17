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

func getIncorrectOptions(word models.Word, level string) []string {
	db := database.DB()
	var incorrectWords []models.Word
	db.Where("level = ? AND english != ?", level, word.English).Order(gorm.Expr("random()")).Limit(3).Find(&incorrectWords)

	var options []string
	for _, incorrectWord := range incorrectWords {
		options = append(options, incorrectWord.Turkish)
	}
	return options
}

func createQuestion(word models.Word, level string) Question {
	options := []string{word.Turkish}
	options = append(options, getIncorrectOptions(word, level)...)

	question := fmt.Sprintf("'%s' kelimesinin Türkçe karşılığı nedir ?", word.English)

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

/*func handleAnswer(bot *tgbotapi.BotAPI, chatID int64, selectedOption string) {
	correctAnswer := correctAnswers[chatID]
	if selectedOption == correctAnswer {
		msg := tgbotapi.NewMessage(chatID, "Doğru!")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Yanlış. Doğru cevap %s.", correctAnswer))
		bot.Send(msg)
	}
	delete(correctAnswers, chatID)
}*/

func handleAnswer(bot *tgbotapi.BotAPI, chatID int64, selectedOption string) bool {
	correctAnswer := correctAnswers[chatID]
	if selectedOption == correctAnswer {
		msg := tgbotapi.NewMessage(chatID, "Doğru!")
		bot.Send(msg)
		delete(correctAnswers, chatID)
		return true
	} else {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Yanlış. Doğru cevap %s.", correctAnswer))
		bot.Send(msg)
		delete(correctAnswers, chatID)
		return false
	}
}
