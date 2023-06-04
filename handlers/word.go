package handlers

import (
	"github.com/huzeyfebostan/go-telegram-bot/database"
	"github.com/huzeyfebostan/go-telegram-bot/models"
)

func AddUserWord(userID int64, englishWord, turkishWord string) error {
	db := database.DB()
	word := models.UserWord{
		UserID:  userID,
		English: englishWord,
		Turkish: turkishWord,
	}
	result := db.Create(&word)
	return result.Error
}

func GetUserWords(userID int64) []models.UserWord {
	var words []models.UserWord
	db := database.DB()
	db.Where("user_id = ?", userID).Find(&words)
	return words
}

func getAllWords(level string) []models.Word {
	var words []models.Word
	db := database.DB()
	db.Where("level = ?", level).Find(&words)
	return words
}

/*func UpdateWord(userID int64, wordID string, english string, turkish string) error {
	var word models.UserWord
	db := database.DB()
	if err := db.Where("id = ? AND user_id = ?", wordID, userID).First(&word).Error; err != nil {
		return err
	}
	word.English = english
	word.Turkish = turkish
	return db.Save(&word).Error
}*/

func UpdateWord(userID int64, originalEnglish string, newEnglish string, newTurkish string) error {
	db := database.DB()

	// First, find the record in the database
	var existingWord models.UserWord
	if err := db.Where("user_id = ? AND English = ?", userID, originalEnglish).First(&existingWord).Error; err != nil {
		return err
	}

	// Then, update the record
	updates := map[string]interface{}{
		"English": newEnglish,
		"Turkish": newTurkish,
	}

	if err := db.Model(&existingWord).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

func DeleteWord(word models.UserWord) error {
	db := database.DB()
	if err := db.Where("English = ?", word.English).First(&word).Error; err != nil {
		return err
	}
	return db.Delete(&word).Error
}
