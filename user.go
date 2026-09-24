package main

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Username       string `gorm:"size:255;uniqueIndex"`
	FirstName      string `gorm:"size:255"`
	LastName       string `gorm:"size:255"`
	HashedPassword string `gorm:"size:255"`
}
