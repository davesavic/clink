package main

import (
	"fmt"
	"github.com/davesavic/clink"
	"net/http"
)

func main() {
	// Create a new client with default options.
	client := clink.NewClient()

	// Create a new request with default options.
	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/anything", nil)

	// Send the request and get the response.
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// Hydrate the response body into a map.
	var target map[string]any
	err = clink.ResponseToJson(resp, &target)

	// Print the target map.
	fmt.Println(target)
}
