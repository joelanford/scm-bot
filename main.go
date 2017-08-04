package main

import (
	"log"

	"github.com/joelanford/scm-bot/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Printf("error: %s", err)
	}
}
