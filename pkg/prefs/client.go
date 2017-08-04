package prefs

import (
	"log"
	"time"

	"github.com/joelanford/scm-bot/pkg/scm"
	"github.com/pkg/errors"
)

type Getter interface {
	Get() (Preferences, error)
}

type Client struct {
	Getter   Getter
	Cloners  map[scm.RepoType]scm.Cloner
	Interval time.Duration
}

func NewClient(getter Getter, interval time.Duration, cloners ...scm.Cloner) *Client {
	clonerMap := make(map[scm.RepoType]scm.Cloner)
	for _, cloner := range cloners {
		clonerMap[cloner.Type()] = cloner
	}
	return &Client{
		Getter:   getter,
		Cloners:  clonerMap,
		Interval: interval,
	}
}

func (c *Client) Run() error {
	ticker := time.NewTicker(c.Interval)
	for {
		log.Println("fetching preferences")
		prefs, err := c.Getter.Get()
		if err != nil {
			log.Println(errors.Wrapf(err, "client: could not get preferences"))
		} else {
			for _, repo := range prefs.GetRepositories() {
				cloner, ok := c.Cloners[repo.Type]
				if !ok {
					log.Println(errors.Errorf("client: no cloner for repo type \"%s\"", repo.Type))
					continue
				}

				if err := cloner.Clone(repo.Path, repo.URL); err != nil {
					log.Println(errors.Wrapf(err, "client: could not clone repo \"%s\" into path \"%s\"", repo.URL, repo.Path))
					continue
				}
				log.Printf("client: cloned repo \"%s\" into path \"%s\"", repo.URL, repo.Path)
			}
		}
		<-ticker.C
	}
}
