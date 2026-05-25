// scan_http.go: Maps opts.HTTP to scanner.HTTPOptions (including default connection timeout and minimum concurrency limit).
package cmd

import (
	"time"

	"swagger-exp-knife4j/pkg/scanner"
)

// scanHTTPOptions returns the HTTP client configuration used by the current CLI scan.
func scanHTTPOptions() *scanner.HTTPOptions {
	connect := opts.HTTP.ConnectTimeout
	if connect == 0 {
		connect = 30 * time.Second
	}
	parallel := opts.HTTP.Parallel
	if parallel < 1 {
		parallel = 1
	}
	return &scanner.HTTPOptions{
		Headers:        opts.HTTP.Headers,
		UserAgent:      opts.HTTP.UserAgent,
		Cookies:        opts.HTTP.Cookies,
		Proxy:          opts.HTTP.Proxy,
		Delay:          opts.HTTP.Delay,
		RequestTimeout: opts.HTTP.RequestTimeout,
		ConnectTimeout: connect,
		Parallel:       parallel,
	}
}
