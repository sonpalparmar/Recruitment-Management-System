package models

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	ApplicantID uint `gorm:"not null"`
	Applicant   User `gorm:"foreignKey:ApplicantID"`
	JobID       uint `gorm:"not null"`
	Job         Job  `gorm:"foreignKey:JobID"`
}
