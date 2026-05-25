// scan_file.go: Implements the `scan file` subcommand.
// Accepts the -f/--file parameter (one URL per line) and invokes scanrun.RunFile for batch scanning.
package cmd

import (
	"errors"
	"fmt"

	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/scanrun"

	"github.com/spf13/cobra"
)

var fileCmdOptions = struct {
	File string
}{}

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Scan multiple targets listed in a file",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if fileCmdOptions.File == "" {
			return errors.New("a targets file must be specified (-f / --file)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		httpOpts := scanHTTPOptions()
		if err := httpOpts.Validate(); err != nil {
			cmd.PrintErrln(err)
			return
		}

		base := scanSingleParams("")
		base.HTTP = httpOpts

		log.InfoStep("Batch scan", "file: "+fileCmdOptions.File)

		batch, err := scanrun.RunFile(fileCmdOptions.File, base, cliBatchHooks())
		if err != nil {
			cmd.PrintErrf("batch scan: %v\n", err)
		}
		if batch == nil {
			return
		}

		log.PrintBatchComplete(fileCmdOptions.File, batch.Total, batch.OK, batch.Failed)

		logBatchWriterOutputs()
		if !scanOutputSatisfiesUser() {
			logNoWritersHint()
		}

		if batch.Failed > 0 && batch.OK == 0 {
			return
		}
		if batch.Failed > 0 {
			log.WarnLine(fmt.Sprintf("batch completed with failures: failed=%d", batch.Failed))
		}
	},
}

func init() {
	SetCommandHelp(fileCmd, helpScanFileLong, helpScanFileExample)
	scanCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&fileCmdOptions.File, "file", "f", "", "A file with target URLs to scan (one per line)")
}
