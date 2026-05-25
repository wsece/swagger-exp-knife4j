// Package cmd provides the command-line entry for swagger-exp-knife4j based on Cobra.
//
// Responsibilities: Parse global and subcommand arguments, assemble pkg/config.Options,
// and invoke modules under pkg/scanrun, pkg/reportserver and pkg/mcpserver.
//
// Subcommand Tree:
//
//	scan single   - Scan a single Swagger JSON or Knife4j/HTML page URL
//	scan file     - Read URLs in bulk from text files and scan them sequentially
//	report list   - List historical scan records from database in terminal table format
//	report server - Launch local web report site with built-in Knife4j debug proxy
//	mcp serve     - Expose MCP tools via stdio for calls from Cursor and other clients
//
// The global variable opts (*config.Options) is declared in root.go,
// shared across all scan and report subcommands.
package cmd
