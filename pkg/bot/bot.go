package bot

import (
	"context"
	"log"
	"path/filepath"

	"github.com/joelanford/scm-bot/pkg/prefs"
	"github.com/joelanford/scm-bot/pkg/scm"
)

type Bot struct {
	baseDirectory string
	getter        prefs.Getter
	cloners       []scm.Cloner
}

func New(baseDirectory string, getter prefs.Getter, cloners ...scm.Cloner) *Bot {
	if absBaseDirectory, err := filepath.Abs(baseDirectory); err == nil {
		baseDirectory = absBaseDirectory
	}
	return &Bot{
		baseDirectory: baseDirectory,
		getter:        getter,
		cloners:       cloners,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	log.Println("bot: started")
	defer log.Println("bot: stopped")

	preferences := b.getter.Get(ctx)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case p, ok := <-preferences:
			if !ok {
				break loop
			}
			if p.Err != nil {
				log.Printf("bot: could not get preferences: %s", p.Err)
			} else if repos := p.Preferences.Repositories; len(repos) == 0 {
				log.Printf("bot: no repositories listed in preferences")
			} else {
				results := scm.CloneAll(b.baseDirectory, b.cloners, repos)
				for result := range results {
					if result.Error == scm.ErrExists {
						log.Printf("bot: skipped cloning repository \"%s\" at path \"%s\": %s", result.URL, result.Path, result.Error)
					} else if result.Error == scm.ErrNoCloner {
						log.Printf("bot: could not clone repo \"%s\": %s", result.URL, result.Error)
					} else if result.Error != nil {
						log.Printf("bot: could not clone repo \"%s\" at path \"%s\": %s", result.URL, result.Path, result.Error)
					} else {
						log.Printf("bot: cloned repo \"%s\" at path \"%s\"", result.URL, result.Path)
					}
				}
			}
		}
	}
	if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
		return ctx.Err()
	}
	return nil
}
