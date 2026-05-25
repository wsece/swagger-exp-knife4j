// tools_register.go: registers swagger_scan, swagger_list_hosts, and swagger_list_sessions on the MCP Server.
package mcpserver

import (
	"swagger-exp-knife4j/pkg/extension"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(s *server.MCPServer, cfg Config) {
	s.AddTool(mcp.NewTool("swagger_scan",
		mcp.WithDescription(toolSwaggerScanDesc),
		mcp.WithString("url", mcp.Required(), mcp.Description(argScanURL)),
		mcp.WithString("output_dir", mcp.Description(argScanOutputDir)),
		mcp.WithBoolean("write_db", mcp.Description(argScanWriteDB)),
		mcp.WithString("db_uri", mcp.Description(argScanDbURI)),
		mcp.WithBoolean("write_csv", mcp.Description(argScanWriteCSV)),
		mcp.WithString("csv_file", mcp.Description(argScanCSVFile)),
		mcp.WithBoolean("write_jsonl", mcp.Description(argScanWriteJSONL)),
		mcp.WithString("jsonl_file", mcp.Description(argScanJSONLFile)),
		mcp.WithNumber("parallel", mcp.Description(argScanParallel)),
	), toolScan(cfg))

	s.AddTool(mcp.NewTool("swagger_list_hosts",
		mcp.WithDescription(toolSwaggerListHostsDesc),
		mcp.WithString("db_uri", mcp.Description(argListHostsDbURI)),
	), toolListHosts(cfg))

	s.AddTool(mcp.NewTool("swagger_list_sessions",
		mcp.WithDescription(toolSwaggerListSessionsDesc),
		mcp.WithString("api_doc_path", mcp.Description(argListSessionsAPIDocPath)),
	), toolListSessions(cfg))

	extension.RegisterMCPToolsOn(s, extension.MCPDefaults{
		DefaultDbURI:      cfg.DefaultDbURI,
		DefaultAPIDocPath: cfg.DefaultAPIDocPath,
	})
}
