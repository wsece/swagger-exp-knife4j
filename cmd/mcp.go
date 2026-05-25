// mcp.go: Implements the `mcp serve` subcommand.
// Runs pkg/mcpserver in stdio mode after silencing logs, for Cursor MCP configuration.
package cmd

import (
	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/mcpserver"

	"github.com/spf13/cobra"
)

var mcpServeOpts = struct {
	DbURI      string
	APIDocPath string
}{}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server for AI clients",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run MCP server on stdio (for Cursor / Claude Desktop)",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.EnableSilence()
		return mcpserver.ServeStdio(mcpserver.Config{
			DefaultDbURI:      mcpServeOpts.DbURI,
			DefaultAPIDocPath: mcpServeOpts.APIDocPath,
		})
	},
}

func init() {
	SetCommandHelp(mcpCmd, helpMCPLong, helpMCPExample)
	SetCommandHelp(mcpServeCmd, helpMCPLong, helpMCPExample)

	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)

	mcpServeCmd.Flags().StringVar(&mcpServeOpts.DbURI, "db-uri", defaultSwaggerDBURI, "Default database URI for tools")
	mcpServeCmd.Flags().StringVar(&mcpServeOpts.APIDocPath, "api-doc-path", "output", "Default api-docs output directory")
}
