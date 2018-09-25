package pget

import (
	"context"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Pget struct {
	parallel int
	timeout time.Duration
}

func NewPget(parallel int, timeout time.Duration) *Pget {
	return &Pget{
		parallel,
		timeout,
	}
}

func (p *Pget) WithCallback(urls []*url.URL, callback func(reader io.Reader) error) error {
	// context with timeout
	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	limit := make(chan struct{}, p.parallel)
	for _, url := range urls {
		eg.Go(func() error {
			// limit parallel executions by using channel
			limit <- struct{}{}
			defer func() { <-limit }()

			select {
			case <-ctx.Done():
				// abort on context cancel
				return nil
			default:
				// get contents
				resp, err := http.Get(url.String())
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				// execute callback
				return callback(resp.Body)
			}
		})
	}
	// wait for all Goroutines
	return eg.Wait()
}
