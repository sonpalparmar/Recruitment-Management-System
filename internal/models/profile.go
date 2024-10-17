package models

import (
	"gorm.io/gorm"
)

type Profile struct {
	gorm.Model
	UserID         uint `gorm:"uniqueIndex;not null"`
	ResumeFilePath string
	Skills         string
	Education      string
	Experience     string
	Name           string
	Email          string
	Phone          string
}
