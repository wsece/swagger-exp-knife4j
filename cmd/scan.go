// scan.go：Implements the `scan` parent command; registers shared PersistentFlags for subcommands including writer, HTTP, and output directory.
package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Swagger API interfaces",
}

func init() {
	SetCommandHelp(scanCmd, helpScanLong, helpScanExample)
	rootCmd.AddCommand(scanCmd)

	fs := scanCmd.PersistentFlags()

	fs.BoolVar(&opts.Writer.Db, "write-db", false, "Write scan results to database (default "+defaultSwaggerDBURI+")")
	fs.StringVar(&opts.Writer.DbURI, "write-db-uri", "", "Database URI (overrides --write-db default)")
	fs.BoolVar(&opts.Writer.DbDebug, "write-db-enable-debug", false, "Show the database query debug logging")

	fs.BoolVar(&opts.Writer.Csv, "write-csv", false, "Write scan results to CSV (default "+defaultScanCSVFile+")")
	fs.StringVar(&opts.Writer.CsvFile, "write-csv-file", "", "CSV file path (overrides --write-csv default)")

	fs.BoolVar(&opts.Writer.Jsonl, "write-jsonl", false, "Write scan results to JSONL (default "+defaultScanJSONLFile+")")
	fs.StringVar(&opts.Writer.JsonlFile, "write-jsonl-file", "", "JSONL file path (overrides --write-jsonl default)")

	fs.StringArrayVarP(&opts.HTTP.Headers, "header", "H", nil, `Custom request header (repeatable), e.g. -H "Authorization: Bearer xxx"`)
	fs.StringVarP(&opts.HTTP.UserAgent, "user-agent", "A", "", `User-Agent string, e.g. -A "Mozilla/5.0"`)
	fs.StringArrayVarP(&opts.HTTP.Cookies, "cookie", "b", nil, `Cookie string or file (repeatable), e.g. -b "session=abc"`)
	fs.StringVarP(&opts.HTTP.Proxy, "proxy", "x", "", `HTTP proxy URL, e.g. -x http://127.0.0.1:8080`)

	fs.DurationVar(&opts.HTTP.Delay, "delay", 0, "Sleep between each API request (e.g. 100ms, 1s)")
	fs.DurationVarP(&opts.HTTP.RequestTimeout, "max-timeout", "m", 0, "Per-request timeout (0 = unlimited)")
	fs.DurationVar(&opts.HTTP.ConnectTimeout, "connect-timeout", 30*time.Second, "Max wait for TCP connect")
	fs.IntVarP(&opts.HTTP.Parallel, "parallel", "P", 1, "Concurrent API request workers")

	fs.StringVar(&opts.Scan.OutputDir, "output-dir", "output", "Base directory for scan output ({host}/{scope}/api-docs.json)")
	fs.BoolVar(&opts.Scan.DocsOnly, "docs-only", false, "Only resolve and dump OpenAPI JSON to --output-dir; skip automated API requests and --write-* outputs")
}
