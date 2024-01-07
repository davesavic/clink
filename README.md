[![test](https://github.com/davesavic/clink/workflows/test/badge.svg)](https://github.com/davesavic/clink/actions?query=workflow%3Atest)
[![coverage](https://coveralls.io/repos/github/davesavic/clink/badge.svg?branch=master)](https://coveralls.io/github/davesavic/clink?branch=master)
[![goreportcard](https://goreportcard.com/badge/github.com/davesavic/clink)](https://goreportcard.com/report/github.com/davesavic/clink)
[![gopkg](https://pkg.go.dev/badge/github.com/davesavic/clink.svg)](https://pkg.go.dev/github.com/davesavic/clink)
[![license](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/davesavic/clink/blob/master/LICENSE)



## Clink: A Configurable HTTP Client for Go

Clink is a highly configurable HTTP client for Go, designed for ease of use, extendability, and robustness. It supports various features like automatic retries and request rate limiting, making it ideal for both simple and advanced HTTP requests.

### Features
- **Flexible Request Options**: Easily configure headers, URLs, and authentication.
- **Retry Mechanism**: Automatic retries with configurable policies.
- **Rate Limiting**: Client-side rate limiting to avoid server-side limits.

### Installation
To use Clink in your Go project, install it using `go get`:

```bash
go get -u github.com/davesavic/clink
```

### Usage
Here is a basic example of how to use Clink:

```go
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
```

*HTTP Methods (HEAD, OPTIONS, GET, HEAD, POST, PATCH, DELETE)* are also supported 
```go
package main

import (
	"github.com/davesavic/clink"
	"encoding/json"
)

func main() {
    client := clink.NewClient()
    resp, err := client.Get("https://httpbin.org/get")
    // ....
    payload, err := json.Marshal(map[string]string{"username": "yumi"})
    resp, err := client.Post("https://httpbin.org/post", payload)
}
```

### Examples
For more examples, see the [examples](https://github.com/davesavic/clink/tree/master/examples) directory.

### Contributing
Contributions to Clink are welcome! If you find a bug, have a feature request, or want to contribute code, please open an issue or submit a pull request.
