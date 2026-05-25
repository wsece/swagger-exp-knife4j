// Package writers persists scan results to external storage (write layer).
//
// Current implementation: SwaggerDbWriter writes scanner.SwaggerReportRecord to GORM
// and groups similar responses via SimHash (reuses internal/islazy.HashGrouper).
package writers
