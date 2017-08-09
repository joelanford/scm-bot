package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/joelanford/scm-bot/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.WithField("component", "main").Fatalln(err)
	}
}
