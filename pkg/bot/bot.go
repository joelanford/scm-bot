package bot

import (
	"context"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/joelanford/scm-bot/pkg/prefs"
	"github.com/joelanford/scm-bot/pkg/scm"
)

type Bot struct {
	baseDirectory string
	getter        prefs.Getter
	cloners       []scm.Cloner
	log           *log.Entry
}

func New(baseDirectory string, getter prefs.Getter, cloners ...scm.Cloner) *Bot {
	if absBaseDirectory, err := filepath.Abs(baseDirectory); err == nil {
		baseDirectory = absBaseDirectory
	}
	return &Bot{
		baseDirectory: baseDirectory,
		getter:        getter,
		cloners:       cloners,
		log:           log.WithFields(log.Fields{"component": "bot"}),
	}
}

func (b *Bot) Run(ctx context.Context) error {
	b.log.Infoln("started")
	defer b.log.Infoln("stopped")

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
				b.log.Warnf("could not get preferences: %s", p.Err)
			} else if repos := p.Preferences.Repositories; len(repos) == 0 {
				b.log.Warnf("no repositories listed in preferences")
			} else {
				results := scm.CloneAll(b.baseDirectory, b.cloners, repos)
				for result := range results {
					if result.Error == scm.ErrExists {
						b.log.Infof("skipped cloning repository \"%s\" at path \"%s\": %s", result.URL, result.Path, result.Error)
					} else if result.Error == scm.ErrNoCloner {
						b.log.Warnf("could not clone repo \"%s\": %s", result.URL, result.Error)
					} else if result.Error != nil {
						b.log.Warnf("could not clone repo \"%s\" at path \"%s\": %s", result.URL, result.Path, result.Error)
					} else {
						b.log.Infof("cloned repo \"%s\" at path \"%s\"", result.URL, result.Path)
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
