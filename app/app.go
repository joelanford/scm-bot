package app

import (
	"os"
	"time"

	"github.com/joelanford/scm-bot/pkg/prefs"
	"github.com/joelanford/scm-bot/pkg/scm"

	"golang.org/x/sync/errgroup"
)

func Run() error {
	getter := &prefs.StaticGetter{
		Preferences: &prefs.StaticPreferences{
			Repositories: []scm.Repository{
				{
					Type: scm.RepoType("git"),
					Path: "gitrepos",
					URL:  "git@github.com:githubtraining/hellogitworld.git",
				},
			},
		},
	}

	baseDir, err := os.Getwd()
	if err != nil {
		return err
	}
	cloner := scm.NewGitCloner(baseDir)

	client := prefs.NewClient(getter, time.Minute*5, cloner)
	server := prefs.NewServer(":8080", cloner)

	var group errgroup.Group
	group.Go(client.Run)
	group.Go(server.Run)

	return group.Wait()
}
