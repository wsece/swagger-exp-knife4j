// server.go: MCP stdio entry, Config, and swagger_scan / list tool handlers.
package mcpserver

import (
	"context"
	"encoding/json"

	"swagger-exp-knife4j/pkg/reportserver"
	"swagger-exp-knife4j/pkg/scanrun"
	"swagger-exp-knife4j/pkg/scanner"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Config holds MCP tool defaults (database URI, api-docs root).
type Config struct {
	DefaultDbURI      string
	DefaultAPIDocPath string
}

// ServeStdio runs the MCP server on stdin/stdout for Cursor / Claude Desktop subprocess use.
func ServeStdio(cfg Config) error {
	if cfg.DefaultDbURI == "" {
		cfg.DefaultDbURI = "sqlite://swagger-scan.sqlite3"
	}
	if cfg.DefaultAPIDocPath == "" {
		cfg.DefaultAPIDocPath = "output"
	}

	s := server.NewMCPServer(
		"swagger-exp-knife4j",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithInstructions(serverInstructions),
		server.WithTitle("Swagger Exp Knife4j MCP"),
		server.WithDescription("Scan Swagger/Knife4j APIs, list DB hosts, list on-disk api-docs sessions. See docs/MCP_TOOLS.md for full tool contracts."),
	)

	registerTools(s, cfg)
	return server.ServeStdio(s)
}

// toolScan handles MCP tool swagger_scan via scanrun.RunSingle (single-scan pipeline).
func toolScan(cfg Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_ = ctx
		url, err := req.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		outputDir := req.GetString("output_dir", cfg.DefaultAPIDocPath)
		parallel := int(req.GetFloat("parallel", 1))
		if parallel < 1 {
			parallel = 1
		}

		w := scanrun.WriterConfig{}
		if req.GetBool("write_db", false) {
			w.DbURI = req.GetString("db_uri", cfg.DefaultDbURI)
		} else if dbURI := req.GetString("db_uri", ""); dbURI != "" {
			w.DbURI = dbURI
		}
		if req.GetBool("write_csv", false) {
			w.CsvFile = req.GetString("csv_file", "result.csv")
		} else if f := req.GetString("csv_file", ""); f != "" {
			w.CsvFile = f
		}
		if req.GetBool("write_jsonl", false) {
			w.JsonlFile = req.GetString("jsonl_file", "result.jsonl")
		} else if f := req.GetString("jsonl_file", ""); f != "" {
			w.JsonlFile = f
		}

		result, err := scanrun.RunSingle(scanrun.SingleParams{
			InputURL:  url,
			OutputDir: outputDir,
			HTTP:      &scanner.HTTPOptions{Parallel: parallel},
			Writers:   w,
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text, err := scanrun.SingleResultJSON(result)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	}
}

// toolListHosts handles MCP tool swagger_list_hosts via reportserver.Store.ListHosts.
// Prerequisite: DB already has records from swagger_scan.
func toolListHosts(cfg Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_ = ctx
		dbURI := req.GetString("db_uri", cfg.DefaultDbURI)
		store, err := reportserver.NewStore(reportserver.Config{
			DbURI:      dbURI,
			APIDocPath: cfg.DefaultAPIDocPath,
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		hosts, err := store.ListHosts()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonToolResult(hosts)
	}
}

// toolListSessions handles MCP tool swagger_list_sessions via reportserver.Store.ListDocSessions.
// Prerequisite: output/ already has api-docs.json from scan (usually after swagger_scan).
func toolListSessions(cfg Config) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_ = ctx
		path := req.GetString("api_doc_path", cfg.DefaultAPIDocPath)
		store, err := reportserver.NewStore(reportserver.Config{
			DbURI:      cfg.DefaultDbURI,
			APIDocPath: path,
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sessions, err := store.ListDocSessions()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonToolResult(sessions)
	}
}

// jsonToolResult returns any JSON-serializable value as a successful tool result (type=text).
func jsonToolResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
