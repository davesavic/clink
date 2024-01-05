package main

import (
	"fmt"
	"github.com/davesavic/clink"
	"net/http"
)

func main() {
	// Create a new client with retries enabled.
	client := clink.NewClient(
		// Retry the request if the status code is 429 (Too Many Requests).
		clink.WithRetries(3, func(req *http.Request, resp *http.Response, err error) bool {
			fmt.Println("Retrying request")

			return resp.StatusCode == http.StatusTooManyRequests
		}),
	)

	// Make a request (randomly selects between status codes 200 and 429).
	for i := 0; i < 10; i++ {
		fmt.Println("Request no.", i)
		req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/status/200%2C429", nil)

		_, err = client.Do(req)
		if err != nil {
			panic(err)
		}
	}
}
