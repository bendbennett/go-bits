package request

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetResult(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
	}))
	defer testServer.Close()

	httpClient := &http.Client{}
	client := NewClient(httpClient.Do, 10)

	resp, err := client.GetResult(context.Background(), testServer.URL)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, nil, err)
}

// TestClient_GetResultChannel uses a mocked server to return a status.
func TestClient_GetResultChannel(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/ok":
			res.WriteHeader(200)
		case "/slow":
			time.Sleep(10 * time.Millisecond)
			res.WriteHeader(200)
		}
	}))
	defer testServer.Close()

	testCases := []struct {
		name           string
		paths          []string
		clientTimeout  time.Duration
		contextTimeout time.Duration
		statusCode     int
		errMsg         string
	}{
		{
			name:           "ok",
			paths:          []string{"/ok", "/ok", "/ok"},
			clientTimeout:  2 * time.Millisecond,
			contextTimeout: 5 * time.Millisecond,
			statusCode:     200,
		},
		{
			name:           "slow - client timeout",
			paths:          []string{"/ok", "/slow", "/ok"},
			clientTimeout:  2 * time.Millisecond,
			contextTimeout: 5 * time.Millisecond,
			statusCode:     200,
			errMsg:         fmt.Sprintf("Get \"%s/slow\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)", testServer.URL),
		},
		{
			name:           "slow - context timeout",
			paths:          []string{"/ok", "/slow", "/ok"},
			clientTimeout:  5 * time.Millisecond,
			contextTimeout: 2 * time.Millisecond,
			statusCode:     200,
			errMsg:         fmt.Sprintf("Get \"%s/slow\": context deadline exceeded", testServer.URL),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := &http.Client{
				Timeout: tc.clientTimeout,
			}

			client := NewClient(httpClient.Do, 10)

			var urls []string
			for _, path := range tc.paths {
				urls = append(urls, fmt.Sprintf("%s%s", testServer.URL, path))
			}

			ctx, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
			defer cancel()
			results := client.GetResultChannel(ctx, urls)

			var resultsCounter int

			for result := range results {
				if result.Err == nil {
					assert.Equal(t, http.StatusOK, result.Status.Code)
				}
				if result.Status == nil {
					assert.Equal(t, tc.errMsg, result.Err.Error())
				}

				resultsCounter++
			}

			assert.Equal(t, len(tc.paths), resultsCounter)
		})
	}
}
