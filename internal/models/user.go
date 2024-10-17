package models

import (
	"gorm.io/gorm"
)

type UserType string

const (
	Admin     UserType = "Admin"
	Applicant UserType = "Applicant"
)

type User struct {
	gorm.Model
	Name            string `gorm:"not null"`
	Email           string `gorm:"uniqueIndex;not null"`
	Address         string
	UserType        UserType `gorm:"type:varchar(10);not null"`
	PasswordHash    string   `gorm:"not null"`
	ProfileHeadline string
	Profile         Profile       `gorm:"foreignKey:UserID"`
	JobsPosted      []Job         `gorm:"foreignKey:PostedByID"`
	Applications    []Application `gorm:"foreignKey:ApplicantID"`
}
