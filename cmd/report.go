// report.go: Implements the `report` parent command.
// Defines the default database URI and the shared --db-uri flag for subcommands.
package cmd

import (
	"github.com/spf13/cobra"
)

const defaultSwaggerDBURI = "sqlite://swagger-scan.sqlite3"

var reportOpts = struct {
	DbURI string
}{}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "View stored Swagger scan reports",
}

func init() {
	SetCommandHelp(reportCmd, helpReportLong, helpReportExample)
	rootCmd.AddCommand(reportCmd)

	reportCmd.PersistentFlags().StringVar(&reportOpts.DbURI, "db-uri", defaultSwaggerDBURI, "Swagger scan database URI (e.g. sqlite://swagger-scan.sqlite3)")
}
