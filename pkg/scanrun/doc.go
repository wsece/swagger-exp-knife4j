// Package scanrun orchestrates single and batch Swagger scan pipelines.
//
// RunSingle: resolve URL → save api-docs → analyze paths → auto-request APIs → optional CSV/JSONL/DB write.
// RunFile: read URLs line-by-line from a file, reuse RunSingle; batch mode supports CSV/JSONL append.
//
// CLI (scan single/file) and MCP (swagger_scan) both call this package to avoid duplicating orchestration in cmd/mcpserver.
package scanrun
