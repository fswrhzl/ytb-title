package main

import (
	"fswrhzl/ytb_title/server/db"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
	if err := db.InitDatabase("server/db/data/ytb_title.db"); err != nil {
		panic(err)
	}
	defer db.Close()
	r := SetupRouter()
	if err := r.Run(":50000"); err != nil {
		panic(err)
	}
}
