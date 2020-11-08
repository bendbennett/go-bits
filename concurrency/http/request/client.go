package request

import (
	"context"
	"net/http"
	"sync"
)

type do func(r *http.Request) (*http.Response, error)

type client struct {
	do               do
	concurrencyLimit int
}

func NewClient(
	do do,
	concurrencyLimit int,
) *client {
	return &client{
		do:               do,
		concurrencyLimit: concurrencyLimit,
	}
}

type result struct {
	Status *status
	Err    error
}

type status struct {
	Code int
	Msg  string
}

func (c *client) GetResult(
	ctx context.Context,
	url string,
) (*status, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)	// 1
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return &status{
		Code: res.StatusCode,
		Msg:  res.Status,
	}, err
}

func (c *client) GetResultChannel(
	ctx context.Context,
	urls []string,
) <-chan result {
	semaphoreChan := make(chan struct{}, c.concurrencyLimit)	// 1
	resultsChan := make(chan result)							// 2

	var wg sync.WaitGroup										// 3
	wg.Add(len(urls))											// 4

	for _, url := range urls {
		go func(url string) {									// 5
			select {
			case <-ctx.Done():									// 6
				wg.Done()
			default:
				semaphoreChan <- struct{}{}						// 7
				resp, err := c.GetResult(ctx, url)				// 8
				resultsChan <- result{
					Status: resp,
					Err:    err,
				}
				<-semaphoreChan									// 9
				wg.Done()										// 10
			}
		}(url)
	}

	go func() {													// 11
		wg.Wait()
		close(resultsChan)
		close(semaphoreChan)
	}()

	return resultsChan
}
