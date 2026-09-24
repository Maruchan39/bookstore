package main

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID              uuid.UUID `gorm:"type:char(36);primaryKey"`
	UserID          uuid.UUID `gorm:"type:char(36);not null;index"`
	User            User      `gorm:"foreignKey:UserID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Title           string `gorm:"size:255"`
	Author          string `gorm:"size:255"`
	PublicationDate time.Time
	Genres          []string `gorm:"serializer:json"`
	IsPrivate       bool
}
