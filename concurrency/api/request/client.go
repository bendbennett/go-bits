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
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
) <-chan *result {
	semaphoreChan := make(chan struct{}, c.concurrencyLimit)
	resultsChan := make(chan *result)

	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, url := range urls {
		go func(url string) {
			select {
			case <-ctx.Done():
				wg.Done()
			default:
				semaphoreChan <- struct{}{}
				resp, err := c.GetResult(ctx, url)
				resultsChan <- &result{
					Status: resp,
					Err:    err,
				}
				<-semaphoreChan
				wg.Done()
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
		close(semaphoreChan)
	}()

	return resultsChan
}
