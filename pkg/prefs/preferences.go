package prefs

import (
	"encoding/json"
	"io"

	"github.com/joelanford/scm-bot/pkg/scm"
	"github.com/pkg/errors"
)

type Preferences interface {
	GetRepositories() []scm.Repository
}

type StaticPreferences struct {
	Repositories []scm.Repository `json:"repositories"`
}

func (p *StaticPreferences) GetRepositories() []scm.Repository {
	return p.Repositories
}

func ReadFrom(r io.Reader) (Preferences, error) {
	d := json.NewDecoder(r)
	var p StaticPreferences
	if err := d.Decode(&p); err != nil {
		return nil, errors.Wrap(err, "could not decode preferences from JSON")
	}
	return &p, nil
}

type StaticPoller struct {
	Preferences *StaticPreferences
}

func (g *StaticPoller) Poll() (Preferences, error) {
	if g.Preferences == nil {
		return nil, errors.New("nil preferences")
	}
	return g.Preferences, nil
}
