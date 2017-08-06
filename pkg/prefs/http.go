package prefs

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type HTTPGetter struct {
	URL       string
	TLSConfig *tls.Config

	init   sync.Once
	client *http.Client
}

func (g *HTTPGetter) Get(ctx context.Context) <-chan GetResult {
	g.init.Do(func() {
		g.client = &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSClientConfig: g.TLSConfig,
			},
		}
	})

	resChan := make(chan GetResult)
	go func() {
		defer close(resChan)
		req, err := http.NewRequest("GET", g.URL, nil)
		if err != nil {
			resChan <- GetResult{nil, errors.Wrapf(err, "could not get preferences from \"%s\"", g.URL)}
			return
		}
		resp, err := g.client.Do(req.WithContext(ctx))
		if err != nil {
			resChan <- GetResult{nil, errors.Wrapf(err, "could not get preferences from \"%s\"", g.URL)}
			return
		}
		defer resp.Body.Close()
		prefs, err := ReadJSON(resp.Body)
		resChan <- GetResult{prefs, err}
	}()

	return resChan
}
