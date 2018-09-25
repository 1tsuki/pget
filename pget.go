package pget

import (
	"context"
	"golang.org/x/sync/errgroup"
	"net/http"
	"net/url"
	"time"
)

type Pget struct {
	timeout time.Duration
	limit chan struct{}
}

func NewPget(parallel int, timeout time.Duration) *Pget {
	return &Pget{
		timeout,
		make(chan struct{}, parallel),
	}
}

func (p *Pget) WithCallback(urls []*url.URL, callback func(url *url.URL, resp *http.Response) error) error {
	// context with timeout
	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	for _, url := range urls {
		p.download(ctx, eg, url, callback)
	}
	return eg.Wait()
}

func (p *Pget) download(ctx context.Context, eg *errgroup.Group, url *url.URL, callback func(*url.URL, *http.Response) error) {
	eg.Go(func() error {
		// limit parallel executions by using channel
		p.limit <- struct{}{}
		defer func() { <-p.limit }()

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
			return callback(url, resp)
		}
	})
}
