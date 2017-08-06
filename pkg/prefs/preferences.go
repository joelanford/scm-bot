package prefs

import (
	"encoding/json"
	"io"

	"github.com/joelanford/scm-bot/pkg/scm"
	"github.com/pkg/errors"
)

type Preferences struct {
	Repositories []scm.Repository `json:"repositories"`
}

func ReadJSON(r io.Reader) (*Preferences, error) {
	d := json.NewDecoder(r)
	var p Preferences
	if err := d.Decode(&p); err != nil {
		return nil, errors.Wrap(err, "could not read JSON preferences")
	}
	return &p, nil
}
