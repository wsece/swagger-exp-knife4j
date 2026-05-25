// scan_output.go: shared CLI logging for scan subcommands.
package cmd

import (
	"fmt"

	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/scanner"
	"swagger-exp-knife4j/pkg/scanrun"
)

func batchProgressLabel(index, total int) string {
	return fmt.Sprintf("[%d/%d]", index, total)
}

func logBatchTargetStart(index, total int, url string) {
	log.InfoStep("Scan target",
		fmt.Sprintf("progress=%s | url=%s", batchProgressLabel(index, total), url),
	)
}

func logBatchTargetDone(index, total int, url string, err error) {
	if err != nil {
		log.InfoStep("Target failed",
			fmt.Sprintf("progress=%s | url=%s", batchProgressLabel(index, total), url),
		)
		return
	}
	log.InfoStep("Target done",
		fmt.Sprintf("progress=%s | url=%s", batchProgressLabel(index, total), url),
	)
}

func cliBatchHooks() *scanrun.BatchHooks {
	return &scanrun.BatchHooks{
		OnTargetStart: logBatchTargetStart,
		OnTargetDone: func(index, total int, url string, result *scanrun.SingleResult, err error) {
			logBatchTargetDone(index, total, url, err)
			if err != nil {
				log.PrintScanFailed(url, err)
				return
			}
			scanner.LogAPIStatisticsDetails(result.Stats)
			printScanSummary(result)
		},
	}
}

func printScanSummary(r *scanrun.SingleResult) {
	if r == nil {
		return
	}
	log.PrintScanFinished(log.ScanFinishedSummary{
		InputURL:     r.InputURL,
		APIs:         r.PathCount,
		RequestOK:    r.RequestOK,
		RequestSkip:  r.RequestSkipped,
		RequestFail:  r.RequestFailed,
		Unauthorized: r.Unauthorized,
		DumpJSON:     r.APIDocsPath,
	})
}

func printScanFailed(url string, err error) {
	log.PrintScanFailed(url, err)
}

func logWriterOutputs(r *scanrun.SingleResult) {
	if r == nil {
		return
	}
	if r.CsvFile != "" {
		log.InfoStep("Write CSV", "file: "+r.CsvFile)
	}
	if r.JsonlFile != "" {
		log.InfoStep("Write JSONL", "file: "+r.JsonlFile)
	}
	if r.WroteDB {
		log.InfoStep("Write database", "uri: "+r.DbURI)
	}
}

func logNoWritersHint() {
	log.WarnLine("no writers have been configured. only print summary. add writers using --write-* flags. " +
		"If you want to view the data report please use --write-db.")
}

func logBatchWriterOutputs() {
	if w := scanWriterConfig(); w.CsvFile != "" {
		log.InfoStep("Write CSV", "file: "+w.CsvFile)
	}
	if w := scanWriterConfig(); w.JsonlFile != "" {
		log.InfoStep("Write JSONL", "file: "+w.JsonlFile)
	}
	if resolvedDbURI() != "" {
		log.InfoStep("Write database", "uri: "+resolvedDbURI())
	}
}
