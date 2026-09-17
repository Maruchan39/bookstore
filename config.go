package main

import "github.com/Maruchan39/bookstore/internal/database"

type Config struct {
	port      string
	dbQueries *database.Queries
}
