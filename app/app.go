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
					Path: os.Getenv("HOME"),
					URL:  "git@github.com:githubtraining/hellogitworld.git",
				},
			},
		},
	}
	cloner := scm.NewGitCloner()

	client := prefs.NewClient(getter, time.Minute*5, cloner)

	// Not doing anything yet...
	server := prefs.NewServer(cloner)

	var group errgroup.Group
	group.Go(client.Run)
	group.Go(server.Run)

	return group.Wait()
}
