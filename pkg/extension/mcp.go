package extension

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPDefaults are passed to extension MCP handlers (same defaults as mcpserver.Config).
type MCPDefaults struct {
	DefaultDbURI      string
	DefaultAPIDocPath string
}

// MCPTool registers an extra MCP tool on `swagger-exp-knife4j mcp serve`.
type MCPTool interface {
	// Name is the tool name exposed to clients (must be unique, snake_case recommended).
	Name() string
	// Definition returns the mcp.Tool schema (description + parameters).
	Definition() mcp.Tool
	// Handler handles tools/call; return mcp.NewToolResultText/Error like built-in tools.
	Handler(defaults MCPDefaults) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// RegisterMCPToolsOn adds all extension MCP tools to s.
func RegisterMCPToolsOn(s *server.MCPServer, defaults MCPDefaults) {
	for _, t := range Default().MCPTools() {
		tool := t.Definition()
		s.AddTool(tool, t.Handler(defaults))
	}
}
