package clink_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/davesavic/clink"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name   string
		opts   []clink.Option
		result func(*clink.Client) bool
	}{
		{
			name: "default client with no options",
			opts: []clink.Option{},
			result: func(client *clink.Client) bool {
				return client.HttpClient != nil && client.Headers != nil && len(client.Headers) == 0
			},
		},
		{
			name: "client with custom http client",
			opts: []clink.Option{
				clink.WithClient(nil),
			},
			result: func(client *clink.Client) bool {
				return client.HttpClient == nil
			},
		},
		{
			name: "client with custom headers",
			opts: []clink.Option{
				clink.WithHeaders(map[string]string{"key": "value"}),
			},
			result: func(client *clink.Client) bool {
				return client.Headers != nil && len(client.Headers) == 1
			},
		},
		{
			name: "client with custom header",
			opts: []clink.Option{
				clink.WithHeader("key", "value"),
			},
			result: func(client *clink.Client) bool {
				return client.Headers != nil && len(client.Headers) == 1
			},
		},
		{
			name: "client with custom rate limit",
			opts: []clink.Option{
				clink.WithRateLimit(60),
			},
			result: func(client *clink.Client) bool {
				return client.RateLimiter != nil && client.RateLimiter.Limit() == 1
			},
		},
		{
			name: "client with basic auth",
			opts: []clink.Option{
				clink.WithBasicAuth("username", "password"),
			},
			result: func(client *clink.Client) bool {
				b64, err := base64.StdEncoding.DecodeString(
					strings.Replace(client.Headers["Authorization"], "Basic ", "", 1),
				)
				if err != nil {
					return false
				}

				return string(b64) == "username:password"
			},
		},
		{
			name: "client with bearer token",
			opts: []clink.Option{
				clink.WithBearerAuth("token"),
			},
			result: func(client *clink.Client) bool {
				return client.Headers["Authorization"] == "Bearer token"
			},
		},
		{
			name: "client with user agent",
			opts: []clink.Option{
				clink.WithUserAgent("user-agent"),
			},
			result: func(client *clink.Client) bool {
				return client.Headers["User-Agent"] == "user-agent"
			},
		},
		{
			name: "client with retries",
			opts: []clink.Option{
				clink.WithRetries(3, func(request *http.Request, response *http.Response, err error) bool {
					return true
				}),
			},
			result: func(client *clink.Client) bool {
				return client.MaxRetries == 3 && client.ShouldRetryFunc != nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := clink.NewClient(tc.opts...)

			if c == nil {
				t.Error("expected client to be created")
			}

			if !tc.result(c) {
				t.Errorf("expected client to be created with options: %+v", tc.opts)
			}
		})
	}
}

func TestClient_Do(t *testing.T) {
	testCases := []struct {
		name        string
		opts        []clink.Option
		setupServer func() *httptest.Server
		resultFunc  func(*http.Response, error) bool
	}{
		{
			name: "successful response no body",
			opts: []clink.Option{},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			resultFunc: func(response *http.Response, err error) bool {
				return response != nil && err == nil && response.StatusCode == http.StatusOK
			},
		},
		{
			name: "successful response with text body",
			opts: []clink.Option{},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("response"))
				}))
			},
			resultFunc: func(response *http.Response, err error) bool {
				bodyContents, err := io.ReadAll(response.Body)
				if err != nil {
					return false
				}

				return string(bodyContents) == "response"
			},
		},
		{
			name: "successful response with json body",
			opts: []clink.Option{},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
				}))
			},
			resultFunc: func(response *http.Response, err error) bool {
				var target map[string]string
				er := clink.ResponseToJson(response, &target)
				if er != nil {
					return false
				}

				return target["key"] == "value"
			},
		},
		{
			name: "successful response with json body and custom headers",
			opts: []clink.Option{
				clink.WithHeaders(map[string]string{"key": "value"}),
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("key") != "value" {
						w.WriteHeader(http.StatusBadRequest)
					}

					_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
				}))
			},
			resultFunc: func(response *http.Response, err error) bool {
				var target map[string]string
				er := clink.ResponseToJson(response, &target)
				if er != nil {
					return false
				}

				return target["key"] == "value"
			},
		},
		{
			name: "successful response with json body and custom header",
			opts: []clink.Option{
				clink.WithHeader("key", "value"),
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("key") != "value" {
						w.WriteHeader(http.StatusBadRequest)
					}

					_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
				}))
			},
			resultFunc: func(response *http.Response, err error) bool {
				var target map[string]string
				er := clink.ResponseToJson(response, &target)
				if er != nil {
					return false
				}

				return target["key"] == "value"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer server.Close()

			opts := append(tc.opts, clink.WithClient(server.Client()))
			c := clink.NewClient(opts...)

			if c == nil {
				t.Error("expected client to be created")
			}

			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			if err != nil {
				t.Errorf("failed to create request: %v", err)
			}

			resp, err := c.Do(req)
			if !tc.resultFunc(resp, err) {
				t.Errorf("expected result to be successful")
			}
		})
	}
}

func TestClient_Methods(t *testing.T) {
	serverFunc := func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Method", r.Method)
		}))
	}
	resultFunc := func(r *http.Response, m string) bool {
		return r.Header.Get("X-Method") == m
	}
	testCases := []struct {
		name        string
		method      string
		body        io.Reader
		setupServer func() *httptest.Server
		resultFunc  func(*http.Response, string) bool
	}{
		{
			name:        "successful head response",
			method:      http.MethodHead,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
		{
			name:        "successful options response",
			method:      http.MethodOptions,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
		{
			name:        "successful get response",
			method:      http.MethodGet,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
		{
			name:        "successful post response",
			method:      http.MethodPost,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
		{
			name:        "successful put response",
			method:      http.MethodPut,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
		{
			name:        "successful patch response",
			method:      http.MethodPatch,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
		{
			name:        "successful delete response",
			method:      http.MethodDelete,
			setupServer: serverFunc,
			resultFunc:  resultFunc,
		},
	}

	call := func(c *clink.Client, method, url string, body io.Reader) (*http.Response, error) {
		switch method {
		case http.MethodHead:
			return c.Head(url)
		case http.MethodOptions:
			return c.Options(url)
		case http.MethodGet:
			return c.Get(url)
		case http.MethodPost:
			return c.Post(url, body)
		case http.MethodPut:
			return c.Put(url, body)
		case http.MethodPatch:
			return c.Patch(url, body)
		case http.MethodDelete:
			return c.Delete(url)
		}
		return nil, nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer server.Close()
			c := clink.NewClient(clink.WithClient(server.Client()))
			if c == nil {
				t.Error("expected client to be created")
			}
			resp, _ := call(c, tc.method, server.URL, tc.body)
			if !tc.resultFunc(resp, tc.method) {
				t.Errorf("expected result to be successful")
			}

		})
	}
}

func TestClient_ResponseToJson(t *testing.T) {
	testCases := []struct {
		name       string
		response   *http.Response
		target     any
		resultFunc func(*http.Response, any) bool
	}{
		{
			name: "successful response with json body",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"key": "value"}`)),
			},
			resultFunc: func(response *http.Response, target any) bool {
				var t map[string]string
				er := clink.ResponseToJson(response, &t)
				if er != nil {
					return false
				}

				return t["key"] == "value"
			},
		},
		{
			name:     "response is nil",
			response: nil,
			resultFunc: func(response *http.Response, target any) bool {
				var t map[string]string
				er := clink.ResponseToJson(response, &t)
				if er == nil {
					return false
				}

				return er.Error() == "response is nil"
			},
		},
		{
			name: "response body is nil",
			response: &http.Response{
				Body: nil,
			},
			resultFunc: func(response *http.Response, target any) bool {
				var t map[string]string
				er := clink.ResponseToJson(response, &t)
				if er == nil {
					return false
				}

				return er.Error() == "response body is nil"
			},
		},
		{
			name: "json decode error",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"key": "value`)),
			},
			target: nil,
			resultFunc: func(response *http.Response, target any) bool {
				var t map[string]string
				er := clink.ResponseToJson(response, &t)
				if er == nil {
					return false
				}

				return strings.Contains(er.Error(), "failed to decode response")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.resultFunc(tc.response, tc.target) {
				t.Errorf("expected result to be successful")
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := clink.NewClient(
		clink.WithRateLimit(60),
		clink.WithClient(server.Client()),
	)

	startTime := time.Now()

	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		if err != nil {
			t.Errorf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("failed to make request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code to be 200")
		}
	}

	elapsedTime := time.Since(startTime)
	if elapsedTime.Seconds() < 0.5 || elapsedTime.Seconds() > 1.5 {
		t.Errorf("expected elapsed time to be between 0.5 and 1.5 seconds, got: %f", elapsedTime.Seconds())
	}
}

func TestSuccessfulRetries(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++ // Increment the request count
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	retryCount := 3
	client := clink.NewClient(
		clink.WithRetries(retryCount, func(request *http.Request, response *http.Response, err error) bool {
			// Check if the response is a 500 Internal Server Error
			return response != nil && response.StatusCode == http.StatusInternalServerError
		}),
		clink.WithClient(server.Client()),
	)

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	if requestCount != retryCount+1 { // +1 for the initial request
		t.Errorf("expected %d retries (total requests: %d), but got %d", retryCount, retryCount+1, requestCount)
	}
}

func TestUnsuccessfulRetries(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++ // Increment the request count
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	retryCount := 3
	client := clink.NewClient(
		clink.WithRetries(retryCount, func(request *http.Request, response *http.Response, err error) bool {
			return false
		}),
		clink.WithClient(server.Client()),
	)

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = client.Do(req)

	if requestCount != 1 { // +1 for the initial request
		t.Errorf("expected %d retries (total requests: %d), but got %d", retryCount, retryCount+1, requestCount)
	}
}
