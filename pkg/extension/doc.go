// Package extension defines registration hooks for third-party modules:
// scan pipeline hooks, CLI subcommands, MCP tools, and optional scan writers.
//
// Register implementations from an extensions/ package init() and blank-import
// that package from main.go. See docs/sphinx/source_md/module-development.md.
package extension
