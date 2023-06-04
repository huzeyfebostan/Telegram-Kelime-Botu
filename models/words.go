package models

import "gorm.io/gorm"

type Word struct {
	gorm.Model
	English string `gorm:"unique;not null"`
	Turkish string `gorm:"not null"`
	Level   string `gorm:"not null"`
}

type UserWord struct {
	gorm.Model
	UserID  int64  `gorm:"not null"`
	English string `gorm:"unique;not null"`
	Turkish string `gorm:"not null"`
}
