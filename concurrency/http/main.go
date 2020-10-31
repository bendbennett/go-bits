package main

import (
	"context"
	"fmt"
	"github.com/bendbennett/go-bits/concurrency/http/request"
	"net/http"
	"runtime"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c := &http.Client{
		Timeout: 450 * time.Millisecond,
	}

	client := request.NewClient(c.Do, runtime.NumCPU())

	urls := []string{
		"https://google.com",
		"https://microsoft.com",
		"https://facebook.com",
	}
	results := client.GetResultChannel(ctx, urls)

	for result := range results {
		if result.Status != nil {
			fmt.Println(result.Status.Msg)
		}
		if result.Err != nil {
			fmt.Println(result.Err)
		}
	}
}
