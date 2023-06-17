package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/huzeyfebostan/go-telegram-bot/database"
	"github.com/huzeyfebostan/go-telegram-bot/models"
	"math/rand"
	"strings"
)

type UserState struct {
	UserID          int64
	SelectedOption  string
	InQuiz          bool
	CurrentWord     models.UserWord
	WordToUpdate    models.UserWord
	WordToDelete    models.UserWord
	CurrentQuestion models.UserWord
	QuestionCount   int
	CorrectAnswer   string
	PreviousMenu    string
}

var userLevels = make(map[int64]string)
var UserPages = make(map[int64]int)
var userMenus = make(map[int64]string)
var userStates = make(map[int64]*UserState)
var questionCount = make(map[int64]int)
var userWords = make(map[int64][]models.UserWord)
var correctCount = make(map[int64]int)
var incorrectCount = make(map[int64]int)

func CreateMainMenu() tgbotapi.InlineKeyboardMarkup {
	button1 := tgbotapi.NewInlineKeyboardButtonData("📖 Seviye Seç", "select_level")
	button2 := tgbotapi.NewInlineKeyboardButtonData("📚 Kelime Listesi", "word_list")
	button3 := tgbotapi.NewInlineKeyboardButtonData("️🗂️ Kelime Yönetimi", "word_management")
	button4 := tgbotapi.NewInlineKeyboardButtonData("📖 Bu Bot Ne Yapabilir ?", "what_bot_can_do")
	button5 := tgbotapi.NewInlineKeyboardButtonData("ℹ️ Hakkında", "about")

	row1 := tgbotapi.NewInlineKeyboardRow(button1)
	row2 := tgbotapi.NewInlineKeyboardRow(button2)
	row3 := tgbotapi.NewInlineKeyboardRow(button3)
	row4 := tgbotapi.NewInlineKeyboardRow(button4)
	row5 := tgbotapi.NewInlineKeyboardRow(button5)

	mainMenu := tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4, row5)

	return mainMenu
}

func createSubMenu() tgbotapi.InlineKeyboardMarkup {
	A1 := tgbotapi.NewInlineKeyboardButtonData("A1", "level_A1")
	A2 := tgbotapi.NewInlineKeyboardButtonData("A2", "level_A2")
	B1 := tgbotapi.NewInlineKeyboardButtonData("B1", "level_B1")
	B2 := tgbotapi.NewInlineKeyboardButtonData("B2", "level_B2")
	C1 := tgbotapi.NewInlineKeyboardButtonData("C1", "level_C1")
	C2 := tgbotapi.NewInlineKeyboardButtonData("C2", "level_C2")
	back := tgbotapi.NewInlineKeyboardButtonData("🔙 Geri", "back")

	row1 := tgbotapi.NewInlineKeyboardRow(A1, A2)
	row2 := tgbotapi.NewInlineKeyboardRow(B1, B2)
	row3 := tgbotapi.NewInlineKeyboardRow(C1, C2)
	row4 := tgbotapi.NewInlineKeyboardRow(back)

	subMenu := tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4)

	return subMenu
}

func createEndQuizMenu() tgbotapi.InlineKeyboardMarkup {
	button1 := tgbotapi.NewInlineKeyboardButtonData("🔁 Tekrar Dene", "retry_quiz")
	button2 := tgbotapi.NewInlineKeyboardButtonData("📖 Seviye Seç", "select_level")
	button3 := tgbotapi.NewInlineKeyboardButtonData("🏠 Ana Menü", "main_menu")

	row := tgbotapi.NewInlineKeyboardRow(button1, button2, button3)

	return tgbotapi.NewInlineKeyboardMarkup(row)
}

func createWordManagementMenu() tgbotapi.InlineKeyboardMarkup {
	button1 := tgbotapi.NewInlineKeyboardButtonData("➕ Kelime Ekle", "add_word")
	button2 := tgbotapi.NewInlineKeyboardButtonData("📚 Kelimelerimi Listele", "user_words")
	button3 := tgbotapi.NewInlineKeyboardButtonData("📝 Kelimelerimden Test Yap", "quiz_words")
	button4 := tgbotapi.NewInlineKeyboardButtonData("🔙 Geri", "back")

	row1 := tgbotapi.NewInlineKeyboardRow(button1)
	row2 := tgbotapi.NewInlineKeyboardRow(button2)
	row3 := tgbotapi.NewInlineKeyboardRow(button3)
	row4 := tgbotapi.NewInlineKeyboardRow(button4)

	return tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4)
}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {

	if _, ok := userStates[query.Message.Chat.ID]; !ok {
		userStates[query.Message.Chat.ID] = &UserState{} // Mevcut değilse kullanıcı için yeni durum
	}

	if query.Data == "select_level" {
		newMenu := createSubMenu()
		msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, newMenu)
		bot.Send(msg)
		userStates[query.Message.Chat.ID].SelectedOption = "select_level" // Kullanıcı durumunu güncelleme
		userStates[query.Message.Chat.ID].PreviousMenu = "main_menu"
	} else if query.Data == "word_management" {
		newMenu := createWordManagementMenu()
		msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, newMenu)
		userStates[query.Message.Chat.ID].PreviousMenu = "main_menu"
		bot.Send(msg)
	} else if query.Data == "add_word" {
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Lütfen İngilizce kelimeyi giriniz:")
		userStates[query.Message.Chat.ID].SelectedOption = "enter_english_word"
		bot.Send(msg)
	} else if query.Data == "user_words" {
		userWords[query.Message.Chat.ID] = GetUserWords(query.Message.Chat.ID)
		UserPages[query.Message.Chat.ID] = 0 // sayfa numarasını sıfırla
		userStates[query.Message.Chat.ID].SelectedOption = "user_words"
		userStates[query.Message.Chat.ID].PreviousMenu = "word_management_menu"
		UserWordPage(bot, query)
	} else if strings.HasPrefix(query.Data, "update_") {
		wordToUpdate := strings.TrimPrefix(query.Data, "update_")
		userStates[query.Message.Chat.ID].SelectedOption = "update_word"
		userStates[query.Message.Chat.ID].WordToUpdate = models.UserWord{English: wordToUpdate}
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Lütfen güncellenmiş İngilizce kelimeyi ve Türkçe karşılığını giriniz: \n ('English - Türkçe' formatında yazınız)")
		bot.Send(msg)
	} else if strings.HasPrefix(query.Data, "delete_") {
		wordToDelete := strings.TrimPrefix(query.Data, "delete_")
		userStates[query.Message.Chat.ID].SelectedOption = "confirm_delete"
		userStates[query.Message.Chat.ID].WordToDelete = models.UserWord{English: wordToDelete}
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Bu kelimeyi silmek istediğinizden emin misiniz?")
		yesButton := tgbotapi.NewInlineKeyboardButtonData("Evet", "yes_delete")
		noButton := tgbotapi.NewInlineKeyboardButtonData("Hayır", "no_delete")
		row := []tgbotapi.InlineKeyboardButton{yesButton, noButton}
		markup := tgbotapi.NewInlineKeyboardMarkup(row)
		msg.ReplyMarkup = &markup
		bot.Send(msg)
	} else if strings.HasPrefix(query.Data, "yes_delete") && userStates[query.Message.Chat.ID].SelectedOption == "confirm_delete" {
		err := DeleteWord(userStates[query.Message.Chat.ID].WordToDelete)
		if err != nil {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Kelimeyi silerken bir hata oluştu. Lütfen tekrar deneyin.")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Kelime başarıyla silindi!")
			bot.Send(msg)
		}
		userStates[query.Message.Chat.ID].WordToDelete = models.UserWord{}
	} else if query.Data == "quiz_words" {
		StartQuiz(bot, query.Message.Chat.ID, query)
	} else if userStates[query.Message.Chat.ID].SelectedOption == "quiz_question" {
		if query.Data == userStates[query.Message.Chat.ID].CorrectAnswer {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Doğru yanıt!")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, fmt.Sprintf("Yanlış yanıt. Doğru yanıt '%s' olacaktı.", userStates[query.Message.Chat.ID].CorrectAnswer))
			bot.Send(msg)
		}
		userStates[query.Message.Chat.ID].QuestionCount++
		if userStates[query.Message.Chat.ID].QuestionCount >= 10 {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Quiz bitti. Tebrikler!")
			bot.Send(msg)
			userStates[query.Message.Chat.ID].PreviousMenu = "word_management_menu"
			retryButton := tgbotapi.NewInlineKeyboardButtonData("🔁 Tekrar Dene", "quiz_words")
			backButton := tgbotapi.NewInlineKeyboardButtonData("🔙 Geri", "back")
			mainMenuButton := tgbotapi.NewInlineKeyboardButtonData("🏠 Ana Menü", "main_menu")

			row := []tgbotapi.InlineKeyboardButton{retryButton, backButton, mainMenuButton}

			markup := tgbotapi.NewInlineKeyboardMarkup(row)

			msg = tgbotapi.NewMessage(query.Message.Chat.ID, "Ne yapmak istersiniz?")
			msg.ReplyMarkup = &markup
			bot.Send(msg)

			userStates[query.Message.Chat.ID].QuestionCount = 0
			userStates[query.Message.Chat.ID].SelectedOption = ""
		} else {
			StartQuiz(bot, query.Message.Chat.ID, query)
		}
	} else if query.Data == "word_list" {
		menu := createSubMenu()
		msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, menu)
		bot.Send(msg)
		userStates[query.Message.Chat.ID].SelectedOption = "word_list"
	} else if query.Data == "back" {
		if userStates[query.Message.Chat.ID].PreviousMenu == "main_menu" {
			mainMenu := CreateMainMenu()
			msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, mainMenu)
			bot.Send(msg)
		} else if userStates[query.Message.Chat.ID].PreviousMenu == "sub_menu" {
			subMenu := createSubMenu()
			msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, subMenu)
			bot.Send(msg)
			userStates[query.Message.Chat.ID].PreviousMenu = "main_menu"
		} else if userStates[query.Message.Chat.ID].PreviousMenu == "word_management_menu" {
			wordManagementMenu := createWordManagementMenu()
			msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, wordManagementMenu)
			bot.Send(msg)
		}
	} else if query.Data == "main_menu" {
		mainMenu := CreateMainMenu()
		msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, mainMenu)
		bot.Send(msg)
	} else if strings.HasPrefix(query.Data, "option_") {
		selectedOption := strings.TrimPrefix(query.Data, "option_")
		//handleAnswer(bot, query.Message.Chat.ID, selectedOption)
		isCorrect := handleAnswer(bot, query.Message.Chat.ID, selectedOption)
		if isCorrect {
			correctCount[query.Message.Chat.ID]++
		} else {
			incorrectCount[query.Message.Chat.ID]++
		}
		questionCount[query.Message.Chat.ID]++
		if questionCount[query.Message.Chat.ID] < 10 {
			if len(userQuestions[query.Message.Chat.ID]) > 0 {
				sendQuestion(bot, query.Message.Chat.ID, userQuestions[query.Message.Chat.ID][0])
				userQuestions[query.Message.Chat.ID] = userQuestions[query.Message.Chat.ID][1:]
			}
		} else {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, fmt.Sprintf("Sınav bitti! %d soruyu doğru, %d soruyu yanlış cevapladınız.", correctCount[query.Message.Chat.ID], incorrectCount[query.Message.Chat.ID]))
			msg.ReplyMarkup = createEndQuizMenu()
			bot.Send(msg)
			questionCount[query.Message.Chat.ID] = 0
			correctCount[query.Message.Chat.ID] = 0
			incorrectCount[query.Message.Chat.ID] = 0
		}
	} else if query.Data == "retry_quiz" {
		questionCount[query.Message.Chat.ID] = 0
		words := getAllWords(userLevels[query.Message.Chat.ID])
		rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
		for _, word := range words[:10] {
			question := createQuestion(word, userLevels[query.Message.Chat.ID])
			userQuestions[query.Message.Chat.ID] = append(userQuestions[query.Message.Chat.ID], question)
		}
		if len(userQuestions[query.Message.Chat.ID]) > 0 {
			sendQuestion(bot, query.Message.Chat.ID, userQuestions[query.Message.Chat.ID][0])
			userQuestions[query.Message.Chat.ID] = userQuestions[query.Message.Chat.ID][1:]
		}
	} else if query.Data == "next_page" {
		NextPage(query.Message.Chat.ID)
		if userStates[query.Message.Chat.ID].SelectedOption == "user_words" {
			UserWordPage(bot, query)
			userStates[query.Message.Chat.ID].PreviousMenu = "word_management_menu"
		} else if userStates[query.Message.Chat.ID].SelectedOption == "word_list" {
			displayPage(bot, query)
		}
	} else if query.Data == "prev_page" {
		PrevPage(query.Message.Chat.ID)
		if userStates[query.Message.Chat.ID].SelectedOption == "user_words" {
			UserWordPage(bot, query)
			userStates[query.Message.Chat.ID].PreviousMenu = "word_management_menu"
		} else if userStates[query.Message.Chat.ID].SelectedOption == "word_list" {
			displayPage(bot, query)
		}
	} else if query.Data == "what_bot_can_do" {
		userStates[query.Message.Chat.ID].PreviousMenu = "main_menu"
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "1.Bu bot, yeni İngilizce kelimeler öğrenmenize yardımcı olabilir. Kelime öğrenmek için öncelikle bir kelime listesi ve seviye seçmeniz gerekmektedir.\n 2.Seçilen seviye ile quiz yapabilir ve yeni ingilizce kelimeler öğrenebilirsiniz.\n 3.Kendi kelime listenizi oluşturabilirsiniz. Kelime listenize kelime eklemek için, kelime yönetimini seçin ve \"kelime ekle\" butonuna tıklayın. Eklemek istediğiniz kelimenin İngilizce ve Türkçe karşılıklarını girmeniz gerekmektedir.\n 4.Kendi kelime listenize göz atabilir ve en az 10 kelime eklediyseniz, kendi kelimelerinizle bir quiz yapabilirsiniz.")
		back := tgbotapi.NewInlineKeyboardButtonData("Geri", "back")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(back))
		bot.Send(msg)
	} else if query.Data == "about" {
		userStates[query.Message.Chat.ID].PreviousMenu = "main_menu"
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Bu bot tarafından @hozeadmin oluşturulmuştur")
		back := tgbotapi.NewInlineKeyboardButtonData("Geri", "back")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(back))
		bot.Send(msg)
	} else {
		levelPrefix := "level_"
		if strings.HasPrefix(query.Data, levelPrefix) {
			userQuestions[query.Message.Chat.ID] = nil
			userLevels[query.Message.Chat.ID] = strings.TrimPrefix(query.Data, levelPrefix)
			if userStates[query.Message.Chat.ID].SelectedOption == "select_level" {
				words := getAllWords(userLevels[query.Message.Chat.ID])
				rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
				for _, word := range words {
					question := createQuestion(word, userLevels[query.Message.Chat.ID])
					userQuestions[query.Message.Chat.ID] = append(userQuestions[query.Message.Chat.ID], question)
				}
				if len(userQuestions[query.Message.Chat.ID]) > 0 {
					sendQuestion(bot, query.Message.Chat.ID, userQuestions[query.Message.Chat.ID][0])
					userQuestions[query.Message.Chat.ID] = userQuestions[query.Message.Chat.ID][1:]
				}
			} else if userStates[query.Message.Chat.ID].SelectedOption == "word_list" {
				if strings.HasPrefix(query.Data, levelPrefix) {
					userLevels[query.Message.Chat.ID] = strings.TrimPrefix(query.Data, levelPrefix)
					UserPages[query.Message.Chat.ID] = 0 // reset the page number
					Words[query.Message.Chat.ID] = GetAllWords(userLevels[query.Message.Chat.ID])
					userMenus[query.Message.Chat.ID] = levelPrefix + userLevels[query.Message.Chat.ID]
					displayPage(bot, query)
				}
			}
		}
		userStates[query.Message.Chat.ID].PreviousMenu = "sub_menu"
	}
}

func GetAllWords(level string) []models.Word {
	var words []models.Word
	db := database.DB()
	db.Where("level = ?", level).Find(&words)
	return words
}

func UserWordPage(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	page := UserPages[query.Message.Chat.ID]
	start := page * 5
	end := min((page+1)*5, len(userWords[query.Message.Chat.ID]))

	if start >= len(userWords[query.Message.Chat.ID]) {
		page = (len(userWords[query.Message.Chat.ID]) / 5) - 1
		start = page * 5
		end = len(userWords[query.Message.Chat.ID])
		UserPages[query.Message.Chat.ID] = page
	}

	if start < 0 {
		page = 0
		start = 0
		end = min(5, len(userWords[query.Message.Chat.ID]))
		UserPages[query.Message.Chat.ID] = 0
	}
	words := userWords[query.Message.Chat.ID][start:end]

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, word := range words {
		button := tgbotapi.NewInlineKeyboardButtonData(word.English+" - "+word.Turkish, "word_"+word.English)
		updateButton := tgbotapi.NewInlineKeyboardButtonData("🔄", "update_"+word.English)
		deleteButton := tgbotapi.NewInlineKeyboardButtonData("❌", "delete_"+word.English)

		row := []tgbotapi.InlineKeyboardButton{button}
		row1 := []tgbotapi.InlineKeyboardButton{updateButton, deleteButton}

		rows = append(rows, row, row1)
	}

	nextButton := tgbotapi.NewInlineKeyboardButtonData("➡️", "next_page")
	prevButton := tgbotapi.NewInlineKeyboardButtonData("⬅️", "prev_page")
	backButton := tgbotapi.NewInlineKeyboardButtonData("🔙 Geri", "back")

	if end == len(userWords[query.Message.Chat.ID]) {
		nextButton = tgbotapi.NewInlineKeyboardButtonData("➡️", "end")
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{prevButton, nextButton, backButton})

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "İşte kelimeleriniz:")
	msg.ReplyMarkup = &markup
	bot.Send(msg)
}

func displayPage(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	page := UserPages[query.Message.Chat.ID]
	start := page * 10
	end := min((page+1)*10, len(Words[query.Message.Chat.ID]))

	if start >= len(Words[query.Message.Chat.ID]) {
		page = (len(Words[query.Message.Chat.ID]) / 10) - 1
		start = page * 10
		end = len(Words[query.Message.Chat.ID])
		UserPages[query.Message.Chat.ID] = page
	}

	if start < 0 {
		page = 0
		start = 0
		end = min(10, len(Words[query.Message.Chat.ID]))
		UserPages[query.Message.Chat.ID] = 0
	}
	words := Words[query.Message.Chat.ID][start:end]

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, word := range words {
		button := tgbotapi.NewInlineKeyboardButtonData(word.English+" - "+word.Turkish, "word_"+word.English)
		row := []tgbotapi.InlineKeyboardButton{button}
		rows = append(rows, row)
	}

	nextButton := tgbotapi.NewInlineKeyboardButtonData("➡️", "next_page")
	prevButton := tgbotapi.NewInlineKeyboardButtonData("⬅️", "prev_page")
	backButton := tgbotapi.NewInlineKeyboardButtonData("🔙 Geri", "back")

	if end == len(Words[query.Message.Chat.ID]) {
		nextButton = tgbotapi.NewInlineKeyboardButtonData("➡️", "end")
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{prevButton, nextButton, backButton})

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Veritabanındaki kelimeler:")
	msg.ReplyMarkup = &markup
	bot.Send(msg)
}

// HandleMessage function
func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Kullanıcı için bir durum olup olmadığını kontrol edin
	if _, ok := userStates[message.Chat.ID]; !ok {
		userStates[message.Chat.ID] = &UserState{}
	}

	if _, ok := userWords[message.Chat.ID]; !ok {
		userWords[message.Chat.ID] = []models.UserWord{} // Eğer mevcut değilse kullanıcı için yeni kelime listesi
	}

	if userStates[message.Chat.ID].SelectedOption == "enter_english_word" {
		userWords[message.Chat.ID] = append(userWords[message.Chat.ID], models.UserWord{English: message.Text})
		msg := tgbotapi.NewMessage(message.Chat.ID, "Lütfen Türkçe çevirisini giriniz:")
		bot.Send(msg)
		userStates[message.Chat.ID].SelectedOption = "enter_turkish_word"
	} else if userStates[message.Chat.ID].SelectedOption == "enter_turkish_word" {
		lastWordIndex := len(userWords[message.Chat.ID]) - 1
		if lastWordIndex >= 0 {
			userWords[message.Chat.ID][lastWordIndex].Turkish = message.Text
			userWords[message.Chat.ID][lastWordIndex].UserID = message.Chat.ID
			err := AddUserWord(message.Chat.ID, userWords[message.Chat.ID][lastWordIndex].English, message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Kelimeyi eklerken bir hata oluştu: "+err.Error())
				bot.Send(msg)
			} else {
				userStates[message.Chat.ID].PreviousMenu = "word_management_menu"
				msg := tgbotapi.NewMessage(message.Chat.ID, "Kelime başarıyla eklendi!")
				backButton := tgbotapi.NewInlineKeyboardButtonData("🔙 Geri", "back")
				mainMenuButton := tgbotapi.NewInlineKeyboardButtonData("🏠 Ana Menü", "main_menu")
				row := tgbotapi.NewInlineKeyboardRow(backButton, mainMenuButton)
				markup := tgbotapi.NewInlineKeyboardMarkup(row)
				msg.ReplyMarkup = &markup
				bot.Send(msg)
			}
			userStates[message.Chat.ID].SelectedOption = ""
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "İngilizce kelimeyi saklarken bir hata oluştu. Lütfen tekrar deneyin.")
			bot.Send(msg)
		}
	} else if userStates[message.Chat.ID].SelectedOption == "update_word" {
		// Parse the user's input
		words := strings.Split(message.Text, " - ")
		if len(words) != 2 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Geçersiz format. Lütfen kelimeyi 'İngilizce - Türkçe' biçiminde girin.")
			bot.Send(msg)
			return
		}

		english := words[0]
		turkish := words[1]

		err := UpdateWord(message.Chat.ID, userStates[message.Chat.ID].WordToUpdate.English, english, turkish)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Kelime güncellenirken bir hata oluştu. Lütfen tekrar deneyin.")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "Word başarıyla güncellendi!")
		bot.Send(msg)
	}

}

func StartQuiz(bot *tgbotapi.BotAPI, chatID int64, query *tgbotapi.CallbackQuery) {
	words := GetUserWords(chatID)

	if len(words) < 10 {
		msg := tgbotapi.NewMessage(chatID, "En az 10 kelime eklemelisiniz.")
		bot.Send(msg)
		return
	}

	rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })

	word := words[0]
	correctAnswer := word.Turkish
	wrongAnswers := GetRandomWords(chatID, 3)

	options := []string{correctAnswer}
	for _, wrongAnswer := range wrongAnswers {
		options = append(options, wrongAnswer.Turkish)
	}

	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	question := fmt.Sprintf("'%s' kelimesinin Türkçe karşılığı nedir?", word.English)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, option := range options {
		button := tgbotapi.NewInlineKeyboardButtonData(option, option)
		row := []tgbotapi.InlineKeyboardButton{button}
		rows = append(rows, row)
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, question)
	msg.ReplyMarkup = &markup
	bot.Send(msg)

	userStates[chatID].CurrentQuestion = word
	userStates[chatID].CorrectAnswer = correctAnswer
	userStates[query.Message.Chat.ID].SelectedOption = "quiz_question"
}

func GetRandomWords(chatID int64, count int) []models.UserWord {
	db := database.DB()
	var words []models.UserWord

	db.Where("user_id = ?", chatID).Find(&words)

	rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })

	return words[:count]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
