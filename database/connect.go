package database

import (
	"github.com/huzeyfebostan/go-telegram-bot/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func DB() *gorm.DB {
	return db
}

func ConnectDB() {
	dbInfo := "host=localhost user=******** password=********** dbname=words_database port=5432 sslmode=disable"

	var err error
	db, err = gorm.Open(postgres.Open(dbInfo), &gorm.Config{})

	if err != nil {
		panic("Veritabanına bağlanılamadı.")
	}

	db.AutoMigrate(models.Word{})
	db.AutoMigrate(models.UserWord{})
}
