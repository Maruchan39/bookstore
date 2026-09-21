package main

import "gorm.io/gorm"

type Config struct {
	port      string
	db        *gorm.DB
	jwtSecret string
}
