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

### Examples
For more examples, see the [examples](https://github.com/davesavic/clink/tree/master/examples) directory.

### Contributing
Contributions to Clink are welcome! If you find a bug, have a feature request, or want to contribute code, please open an issue or submit a pull request.