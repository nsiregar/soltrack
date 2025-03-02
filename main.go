/*
Copyright © 2025 Ngalim Siregar ngalim.siregar@gmail.com
*/
package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/nsiregar/soltrack/cmd"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}
	cmd.Execute()
}
