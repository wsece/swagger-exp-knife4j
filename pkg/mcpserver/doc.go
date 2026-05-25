// Package mcpserver exposes swagger-exp-knife4j as an MCP (Model Context Protocol) stdio server.
//
// File layout:
//   - doc.go: package overview (this file)
//   - server.go: ServeStdio and tool handlers
//   - tools_register.go: register tools on the mcp-go Server
//   - tools_schema.go: tool names, descriptions, and parameter schema constants
//
// # When to enable
//
// Start as a subprocess in Cursor / Claude Desktop (or similar) config:
//
//	command: swagger-exp-knife4j.exe
//	args: ["mcp", "serve"]
//
// The client talks over stdio; the model invokes capabilities via tools/list and tools/call.
//
// # Tools (call order see serverInstructions)
//
//   - swagger_scan        — scan one Swagger/Knife4j target (primary write path)
//   - swagger_list_hosts  — list scanned hosts from DB (needs prior scan with write_db)
//   - swagger_list_sessions — list on-disk api-docs sessions (needs output/ after scan)
//
// # Contract details
//
// Per-tool purpose, parameter types, return JSON, prerequisites, and recommended flows
// live in tools_schema.go constants and are injected into MCP tool descriptions and server instructions.
//
// Human-readable copy: docs/mcp-tools.md at repo root.
package mcpserver
