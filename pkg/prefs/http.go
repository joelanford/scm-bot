package prefs

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type HTTPGetter struct {
	PreferencesURL string
	HTTPClient     *http.Client
}

func (g *HTTPGetter) Get() (Preferences, error) {
	if g.HTTPClient == nil {
		g.HTTPClient = &http.Client{
			Timeout: time.Second * 10,
		}
	}
	resp, err := g.HTTPClient.Get(g.PreferencesURL)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get preferences from \"%s\"", g.PreferencesURL)
	}
	defer resp.Body.Close()

	return ReadFrom(resp.Body)
}
