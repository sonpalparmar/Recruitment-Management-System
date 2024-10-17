package models

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Title             string        `gorm:"not null"`
	Description       string        `gorm:"not null"`
	PostedOn          time.Time     `gorm:"autoCreateTime"`
	TotalApplications int           `gorm:"default:0"`
	CompanyName       string        `gorm:"not null"`
	PostedByID        uint          `gorm:"not null"`
	PostedBy          User          `gorm:"foreignKey:PostedByID"`
	Applications      []Application `gorm:"foreignKey:JobID"`
}
