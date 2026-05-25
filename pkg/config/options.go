// options.go: Defines the global CLI configuration struct (opts is populated after parsing flags in cmd).
package config

import "time"

// Options aggregates all configurable CLI options of swagger-exp-knife4j.
type Options struct {
	Logging Logging // Logging: Debug / Silent
	Writer  Writer  // Scan result output: DB / CSV / JSONL
	Scan    Scan    // Scan behavior: output directory, etc.
	HTTP    HTTP    // HTTP probe: headers, cookies, proxy, timeouts, concurrency
}

// HTTP holds curl-style request parameters (scan flags -H/-A/-b/-x, etc.).
type HTTP struct {
	Headers        []string      // Custom headers, format "Key: Value"
	UserAgent      string        // User-Agent; -A/--user-agent
	Cookies        []string      // Cookie strings or file paths; -b/--cookie
	Proxy          string        // HTTP proxy URL; -x/--proxy
	Delay          time.Duration // Delay between each API request; --delay
	RequestTimeout time.Duration // Per-request total timeout; 0 means no limit; -m/--max-timeout
	ConnectTimeout time.Duration // TCP connect timeout; --connect-timeout
	Parallel       int           // Concurrent worker count; -P/--parallel
}

// Logging controls pkg/log output level.
type Logging struct {
	Debug   bool // -D/--debug-log
	Silence bool // -q/--quiet
}

// Writer controls where scan results are persisted.
type Writer struct {
	Db        bool   // --write-db
	DbURI     string // --write-db-uri, overrides default sqlite URI
	DbDebug   bool   // --write-db-enable-debug
	Csv       bool   // --write-csv
	CsvFile   string // --write-csv-file
	Jsonl     bool   // --write-jsonl
	JsonlFile string // --write-jsonl-file
}

// Scan holds scan-run options.
type Scan struct {
	OutputDir string // Root directory for api-docs.json; --output-dir, default output
	DocsOnly  bool   // --docs-only: dump OpenAPI JSON only, skip automated API requests
}
