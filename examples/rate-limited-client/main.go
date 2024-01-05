package main

import (
	"fmt"
	"github.com/davesavic/clink"
	"net/http"
)

func main() {
	// Create a new client with a limit of 60 requests per minute (1 per second).
	client := clink.NewClient(
		clink.WithRateLimit(60),
	)

	// Create a new request with default options.
	req, _ := http.NewRequest(http.MethodGet, "https://httpbin.org/anything", nil)

	reqCount := 0
	for i := 0; i < 100; i++ {
		fmt.Println("Request no.", i)
		reqCount++

		// Send the rate limited request and get the response.
		// The client will wait for the rate limiter to allow the request.
		_, err := client.Do(req)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(reqCount)
}
