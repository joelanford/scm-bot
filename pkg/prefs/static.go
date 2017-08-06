package prefs

import (
	"context"
)

type StaticGetter struct {
	Preferences Preferences
}

func (g *StaticGetter) Get(context.Context) <-chan GetResult {
	resChan := make(chan GetResult)

	go func() {
		resChan <- GetResult{&g.Preferences, nil}
		close(resChan)
	}()

	return resChan
}
