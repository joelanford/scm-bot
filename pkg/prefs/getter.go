package prefs

import (
	"context"
	"sync"
)

type Getter interface {
	Get(context.Context) <-chan GetResult
}

type GetResult struct {
	Preferences *Preferences
	Err         error
}

type getter struct {
	getters []Getter
}

var defaultGetter getter

func (g *getter) Get(ctx context.Context) <-chan GetResult {
	resChan := make(chan GetResult)

	var wg sync.WaitGroup
	for _, getter := range g.getters {
		wg.Add(1)
		go func(c <-chan GetResult) {
			for res := range c {
				resChan <- res
			}
			wg.Done()
		}(getter.Get(ctx))
	}
	go func() {
		wg.Wait()
		close(resChan)
	}()

	return resChan
}

func UseGetter(g Getter) {
	defaultGetter.getters = append(defaultGetter.getters, g)
}

func Get(ctx context.Context) <-chan GetResult {
	return defaultGetter.Get(ctx)
}

func DefaultGetter() Getter {
	return &defaultGetter
}
