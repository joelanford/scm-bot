package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/joelanford/scm-bot/pkg/bot"
	"github.com/joelanford/scm-bot/pkg/prefs"
	"github.com/joelanford/scm-bot/pkg/scm"
)

var (
	appName   string
	version   string
	buildUser string
	gitHash   string

	cloners []scm.Cloner

	log = logrus.WithField("component", "app")
)

func init() {
	cloners = append(cloners, scm.NewGitCloner())
}

func printVersion(c *cli.Context) {
	fmt.Printf("Version:     %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildUser, gitHash)
}

func Run() error {
	cli.OsExiter = func(_ int) {}
	cli.ErrWriter = ioutil.Discard
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = appName
	app.HelpName = appName
	app.Usage = "Automatically clone source code repositories"

	app.Version = version

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
		cli.StringFlag{
			Name:  "log-level,l",
			Usage: "Log level (one of panic, fatal, error, warn, info, debug)",
			Value: "info",
		},
		cli.StringFlag{
			Name:  "log-fmt, f",
			Usage: "Log format (one of text, json)",
			Value: "text",
		},
	}

	app.Before = before

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

func before(c *cli.Context) error {
	err1 := setLogFormat(c.GlobalString("log-fmt"))
	err2 := setLogLevel(c.GlobalString("log-level"))

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}
	return logGlobalFlags(c)
}

func setLogLevel(logLevel string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	return nil
}

func setLogFormat(logFmt string) error {
	switch logFmt {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		return fmt.Errorf("unrecognized log format: \"%s\"", logFmt)
	}
	return nil
}

func logGlobalFlags(c *cli.Context) error {
	for _, f := range c.GlobalFlagNames() {
		log.WithFields(logrus.Fields{
			"scope":     "global",
			"flagName":  f,
			"flagValue": c.GlobalGeneric(f),
		}).Info("global flag set")
	}
	return nil
}

func logCommandFlags(c *cli.Context) error {
	for _, f := range c.FlagNames() {
		log.WithFields(logrus.Fields{
			"scope":     "command",
			"flagName":  f,
			"flagValue": c.Generic(f),
		}).Info("command flag set")
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
		return errors.New("--url flag must be defined")
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
		return errors.New("--url flag must be defined")
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
