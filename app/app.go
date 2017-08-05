package app

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joelanford/scm-bot/pkg/prefs"
	"github.com/joelanford/scm-bot/pkg/scm"
	"github.com/urfave/cli"
)

var (
	appName   string
	version   string
	buildTime string
	buildUser string
	gitHash   string

	cloners []scm.Cloner
)

func init() {
	cloners = append(cloners, scm.NewGitCloner())
}

func PrintVersion(ctx *cli.Context) {
	fmt.Printf("Version:     %s\nBuild Time:  %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildTime, buildUser, gitHash)
}

func Run() error {
	cli.VersionPrinter = PrintVersion
	app := cli.NewApp()

	app.Name = appName
	app.HelpName = appName
	app.Version = version
	if compiled, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", buildTime); err == nil {
		app.Compiled = compiled
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "basedir, b",
			Value: filepath.Join(os.Getenv("HOME"), "scm-bot"),
			Usage: "`BASEDIR` in which scm-bot will clone repositories",
		},
		cli.DurationFlag{
			Name:  "interval, i",
			Value: time.Minute * 1,
			Usage: "`INTERVAL` of scm-bot client preferences scrapes to poll for updates",
		},
	}

	app.Before = logGlobalFlags

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "http",
			Before: logCommandFlags,
			Action: pollHttp,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "url, u",
					Usage: "`URL` at which scm-bot can scrape user preferences",
				},
				cli.StringFlag{
					Name:  "tls-client-cert, c",
					Usage: "`FILE` containing client certificate, if required by preferences server",
				},
				cli.StringFlag{
					Name:  "tls-client-key, k",
					Usage: "`FILE` containing client certificate key, if required by preferences server",
				},
			},
		},
		cli.Command{
			Name:   "static",
			Before: logCommandFlags,
			Action: pollStatic,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path, p",
					Usage: "`PATH` under base directory at which to clone static repository",
					Value: ".",
				},
				cli.StringFlag{
					Name:  "type, t",
					Usage: "`TYPE` of repository to scrape",
					Value: "git",
				},
				cli.StringFlag{
					Name:  "url, u",
					Usage: "`URL` of repository to clone",
				},
			},
		},
	}
	return app.Run(os.Args)
}

func logGlobalFlags(ctx *cli.Context) error {
	for _, f := range ctx.GlobalFlagNames() {
		log.Printf("FLAG --%s=%s", f, ctx.Generic(f))
	}
	return nil
}

func logCommandFlags(ctx *cli.Context) error {
	for _, f := range ctx.FlagNames() {
		log.Printf("FLAG --%s=%s", f, ctx.Generic(f))
	}
	return nil
}

func pollHttp(ctx *cli.Context) error {
	baseDir := ctx.GlobalString("basedir")
	interval := ctx.GlobalDuration("interval")

	url := ctx.String("url")
	certFile := ctx.String("tls-client-cert")
	keyFile := ctx.String("tls-client-cert-key")
	insecureSkipVerify := ctx.IsSet("tls-insecure-skip-verify")

	if url == "" {
		return cli.NewExitError("ERROR: \"--url\" must be defined", 1)
	}

	tlsConf := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: insecureSkipVerify,
	}
	if certFile != "" {
		var (
			cert tls.Certificate
			err  error
		)
		if keyFile != "" {
			cert, err = tls.LoadX509KeyPair(certFile, keyFile)
		} else {
			cert, err = tls.LoadX509KeyPair(certFile, certFile)
		}
		if err != nil {
			return err
		}
		tlsConf.Certificates = []tls.Certificate{cert}
	}

	poller := &prefs.HTTPPoller{URL: url, TLSConfig: tlsConf}
	client := prefs.NewPollClient(baseDir, poller, interval, cloners...)
	return client.Run()
}

func pollStatic(ctx *cli.Context) error {
	baseDir := ctx.GlobalString("basedir")
	interval := ctx.GlobalDuration("interval")

	repoURL := ctx.String("url")
	repoType := ctx.String("type")
	repoPath := ctx.String("path")

	if repoURL == "" {
		return cli.NewExitError("ERROR: \"--url\" must be defined", 1)
	}

	poller := &prefs.StaticPoller{
		Preferences: &prefs.StaticPreferences{
			Repositories: []scm.Repository{
				{
					URL:  repoURL,
					Type: scm.RepoType(repoType),
					Path: repoPath,
				},
			},
		},
	}
	client := prefs.NewPollClient(baseDir, poller, interval, cloners...)
	return client.Run()
}
