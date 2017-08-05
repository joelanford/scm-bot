package prefs

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type HTTPPoller struct {
	URL       string
	TLSConfig *tls.Config

	init   sync.Once
	client *http.Client
}

func (p *HTTPPoller) Poll() (Preferences, error) {
	p.init.Do(func() {
		p.client = &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSClientConfig: p.TLSConfig,
			},
		}
	})

	resp, err := p.client.Get(p.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get preferences from \"%s\"", p.URL)
	}
	defer resp.Body.Close()

	return ReadFrom(resp.Body)
}
