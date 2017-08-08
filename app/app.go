package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joelanford/scm-bot/pkg/bot"
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

func PrintVersion(c *cli.Context) {
	fmt.Printf("Version:     %s\nBuild Time:  %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildTime, buildUser, gitHash)
}

func Run() error {
	cli.VersionPrinter = PrintVersion
	app := cli.NewApp()

	app.Name = appName
	app.HelpName = appName
	app.Version = version
	if compiled, err := time.Parse("2006-01-02 15:04:05 -0700 MST", buildTime); err == nil {
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
			Usage: "If set, request preferences every `INTERVAL`",
		},
		cli.DurationFlag{
			Name:  "duration, d",
			Usage: "If set, stop scm-bot after `DURATION`",
		},
	}

	app.Before = logGlobalFlags

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "http",
			Action: runHttp,
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
			Action: runStatic,
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

func logGlobalFlags(c *cli.Context) error {
	for _, f := range c.GlobalFlagNames() {
		log.Printf("GLOBAL FLAG --%s=%s", f, c.GlobalGeneric(f))
	}
	return nil
}

func logCommandFlags(c *cli.Context) error {
	for _, f := range c.FlagNames() {
		log.Printf("FLAG --%s=%s", f, c.Generic(f))
	}
	return nil
}

func runHttp(c *cli.Context) error {
	logCommandFlags(c)

	baseDir := c.GlobalString("basedir")
	interval := c.GlobalDuration("interval")
	duration := c.GlobalDuration("duration")

	url := c.String("url")
	certFile := c.String("tls-client-cert")
	keyFile := c.String("tls-client-cert-key")
	insecureSkipVerify := c.IsSet("tls-insecure-skip-verify")

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

	getter := prefs.HTTPGetter{
		URL:       url,
		TLSConfig: tlsConf,
	}
	return runBot(&getter, baseDir, interval, duration)
}

func runStatic(c *cli.Context) error {
	logCommandFlags(c)

	baseDir := c.GlobalString("basedir")
	interval := c.GlobalDuration("interval")
	duration := c.GlobalDuration("duration")

	repoURL := c.String("url")
	repoType := c.String("type")
	repoPath := c.String("path")

	if repoURL == "" {
		return cli.NewExitError("ERROR: \"--url\" must be defined", 1)
	}

	getter := prefs.StaticGetter{
		Preferences: prefs.Preferences{
			Repositories: []scm.Repository{
				{
					URL:  repoURL,
					Type: scm.RepoType(repoType),
					Path: repoPath,
				},
			},
		},
	}
	return runBot(&getter, baseDir, interval, duration)
}

func runBot(getter prefs.Getter, baseDir string, interval, duration time.Duration) error {
	if interval > 0 {
		getter = prefs.OnInterval(getter, interval)
	}
	prefs.UseGetter(getter)

	b := bot.New(baseDir, prefs.DefaultGetter(), cloners...)

	ctx := context.Background()
	if duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}

	return b.Run(ctx)
}
