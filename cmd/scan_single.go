// scan_single.go: Implements the `scan single` subcommand.
// Accepts the -u/--url parameter, invokes scanrun.RunSingle, and prints scan statistics.
package cmd

import (
	"errors"

	"swagger-exp-knife4j/pkg/scanrun"

	"github.com/spf13/cobra"
)

var singleCmdOptions = struct {
	URL string
}{}

var singleCmd = &cobra.Command{
	Use:   "single",
	Short: "Scan a single URL target",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if singleCmdOptions.URL == "" {
			return errors.New("a URL must be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		httpOpts := scanHTTPOptions()
		if err := httpOpts.Validate(); err != nil {
			printScanFailed(singleCmdOptions.URL, err)
			return
		}

		result, err := scanrun.RunSingle(scanSingleParams(singleCmdOptions.URL))
		if err != nil {
			printScanFailed(singleCmdOptions.URL, err)
			return
		}

		printScanSummary(result)
		logWriterOutputs(result)
		if !scanOutputSatisfiesUser() {
			logNoWritersHint()
		}
	},
}

func init() {
	SetCommandHelp(singleCmd, helpScanSingleLong, helpScanSingleExample)
	scanCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&singleCmdOptions.URL, "url", "u", "", "Swagger JSON or Swagger UI / Knife4j HTML page URL")
}
