// report_server.go: Implements the `report server` subcommand.
// Starts the local Web server in pkg/reportserver (default: 127.0.0.1:7171).
package cmd

import (
	"fmt"

	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/reportserver"

	"github.com/spf13/cobra"
)

var serverCmdFlags = struct {
	Host         string
	Port         int
	DbURI        string
	APIDocPath   string
	Knife4jProxy bool
}{}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start local web server to browse scan results",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := reportserver.NewStore(reportserver.Config{
			DbURI:      serverCmdFlags.DbURI,
			APIDocPath: serverCmdFlags.APIDocPath,
		})
		if err != nil {
			return fmt.Errorf("open store: %w", err)
		}
		hosts, _ := store.ListHosts()
		sessions, _ := store.ListDocSessions()
		log.Print("report server data loaded",
			"hosts", len(hosts),
			"sessions", len(sessions),
			"db", serverCmdFlags.DbURI,
			"output", store.Config().APIDocPath,
		)

		srv := reportserver.NewServer(store, reportserver.ServerConfig{
			Host:         serverCmdFlags.Host,
			Port:         serverCmdFlags.Port,
			Knife4jProxy: serverCmdFlags.Knife4jProxy,
		})
		if err := srv.ListenAndServe(); err != nil {
			return formatReportListenError(serverCmdFlags.Host, serverCmdFlags.Port, err)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	SetCommandHelp(serverCmd, helpReportServerLong, helpReportServerExample)
	reportCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVar(&serverCmdFlags.Host, "host", "127.0.0.1", "The host address to bind the webserver to")
	serverCmd.Flags().IntVar(&serverCmdFlags.Port, "port", 7171, "The port to start the web server on")
	serverCmd.Flags().StringVar(&serverCmdFlags.DbURI, "db-uri", defaultSwaggerDBURI, "The database URI for scan records")
	serverCmd.Flags().StringVar(&serverCmdFlags.APIDocPath, "api-doc-path", "./output", "The path where api-docs.json files are stored")
	serverCmd.Flags().BoolVar(&serverCmdFlags.Knife4jProxy, "knife4j-proxy", false, "Route Knife4j try-it-out via same-origin proxy (CORS workaround); default sends requests to target host for Burp etc.")
}
