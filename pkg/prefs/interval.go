package prefs

import (
	"context"
	"time"
)

type IntervalGetter struct {
	Getter   Getter
	Interval time.Duration
}

func OnInterval(g Getter, interval time.Duration) Getter {
	return &IntervalGetter{g, interval}
}

func (g *IntervalGetter) Get(ctx context.Context) <-chan GetResult {
	resChan := make(chan GetResult)
	go func() {
		defer close(resChan)

		getGetter := func() {
			for p := range g.Getter.Get(ctx) {
				resChan <- p
			}
		}
		ticker := time.NewTicker(g.Interval)

		select {
		case <-ctx.Done():
			return
		default:
			getGetter()
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				getGetter()
			}
		}
	}()
	return resChan
}
