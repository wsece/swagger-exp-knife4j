// help_content.go provides Long help texts and Example usage for all CLI commands.
package cmd

const prog = "swagger-exp-knife4j"

const (
	helpScanLong = `## This module[scan] scans Swagger/Knife4j/OpenAPI targets and auto-detect interfaces. 
## Subcommands: single (single URL), file (URL list file). 
## Flags for write config, HTTP options, output directory and more are shown below.`

	helpScanExample = prog + ` scan single -u https://example.com/doc.html --write-db
` + prog + ` scan file -f targets.txt --write-db`

	helpScanSingleLong = `## This subcommand[scan][single] scan a single URL (Swagger JSON, Knife4j or Swagger UI page).
## Use with parent scan command flags such as --write-db, -H, -P, see scan --help for details.`

	helpScanSingleExample = prog + ` scan single -u https://example.com/doc.html --write-db`

	helpScanFileLong = `## This subcommand[scan][file] reads URL list files and executes scanning.`

	helpScanFileExample = prog + ` scan file -f targets.txt --write-db`

	helpReportLong = `## This module[report] enables report viewing and displays scanned results.。`

	helpReportExample = prog + ` report list
` + prog + ` report server`

	helpReportListLong = `## This subcommand[report][list] List API detection records from the database (terminal table format).。`

	helpReportListExample = prog + ` report list`

	helpReportServerLong = `## This subcommand[report][server] Start the web service to view results and dispatch scanning tasks.`

	helpReportServerExample = prog + ` report server
` + prog + ` report server --api-doc-path ./output --db-uri sqlite://swagger-scan.sqlite3`

	helpMCPLong = `## This module[mcp] is used to start the MCP service for invocation by large language models.`

	helpMCPExample = prog + ` mcp serve`

	helpVersionLong = `## This subcommand[version] prints release version and build metadata (from internal/version).`

	helpVersionExample = prog + ` version`
)
