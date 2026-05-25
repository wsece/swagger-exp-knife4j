// Package config defines the CLI global configuration struct (library layer, no business logic).
//
// The cmd package populates Options after parsing Cobra flags, then distributes them to pkg/scanrun and pkg/scanner.
// Includes: log level, scan output directory, HTTP request customization, result output (database/CSV/JSONL).
package config
