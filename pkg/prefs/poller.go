package prefs

import (
	"log"
	"path/filepath"
	"time"

	"github.com/joelanford/scm-bot/pkg/scm"
)

type Poller interface {
	Poll() (Preferences, error)
}

type PollClient struct {
	BaseDirectory string
	Poller        Poller
	Cloners       []scm.Cloner
	Interval      time.Duration
}

func NewPollClient(baseDirectory string, poller Poller, interval time.Duration, cloners ...scm.Cloner) *PollClient {
	if absBaseDirectory, err := filepath.Abs(baseDirectory); err == nil {
		baseDirectory = absBaseDirectory
	}
	return &PollClient{
		BaseDirectory: baseDirectory,
		Poller:        poller,
		Cloners:       cloners,
		Interval:      interval,
	}
}

func (c *PollClient) Run() error {
	ticker := time.NewTicker(c.Interval)
	for {
		log.Println("client: fetching preferences")
		if prefs, err := c.Poller.Poll(); err != nil {
			log.Printf("client: could not get preferences: %s", err)
		} else if repos := prefs.GetRepositories(); len(repos) == 0 {
			log.Printf("client: no repositories listed in preferences")
		} else {
			results := scm.CloneAll(c.BaseDirectory, c.Cloners, repos)
			for result := range results {
				if result.Error == scm.ErrExists {
					log.Printf("client: skipped cloning repository \"%s\" at path \"%s\": %s", result.URL, result.Path, result.Error)
				} else if result.Error == scm.ErrNoCloner {
					log.Printf("client: could not clone repo \"%s\": %s", result.URL, result.Error)
				} else if result.Error != nil {
					log.Printf("client: could not clone repo \"%s\" at path \"%s\": %s", result.URL, result.Path, result.Error)
				} else {
					log.Printf("client: cloned repo \"%s\" at path \"%s\"", result.URL, result.Path)
				}
			}
		}
		<-ticker.C
	}
}
