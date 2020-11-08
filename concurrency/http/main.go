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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) // 1
	defer cancel()

	c := &http.Client{
		Timeout: 450 * time.Millisecond, // 2
	}

	client := request.NewClient(c.Do, runtime.NumCPU()) // 3

	urls := []string{
		"https://google.com",
		"https://microsoft.com",
		"https://facebook.com",
	}
	results := client.GetResultChannel(ctx, urls) // 4

	for result := range results { // 5
		if result.Status != nil {
			fmt.Println(result.Status.Msg)
		}
		if result.Err != nil {
			fmt.Println(result.Err)
		}
	}
}
