// cli_format.go: structured scan CLI messages via charmbracelet log levels.
package log

import (
	"fmt"
	"strings"

	"swagger-exp-knife4j/internal/islazy"
)

const stepLabelWidth = 16

func formatStep(label, detail string) string {
	if detail == "" {
		return label
	}
	return fmt.Sprintf("%-*s | %s", stepLabelWidth, label, detail)
}

// InfoStep logs one pipeline step at Info level (hidden with -q / EnableSilence).
func InfoStep(label, detail string) {
	Logger.Helper()
	Logger.Info(formatStep(label, detail))
}

// InfoStepf formats detail then logs InfoStep.
func InfoStepf(label, format string, args ...interface{}) {
	InfoStep(label, fmt.Sprintf(format, args...))
}

// WarnLine logs a warning at Warn level (hidden with -q).
func WarnLine(msg string) {
	Logger.Helper()
	Logger.Warn(strings.TrimSpace(msg))
}

// ScanFinishedSummary is the per-target outcome block (Print level, visible with -q).
// Set Result for failure text; leave empty on success to format API/request stats.
type ScanFinishedSummary struct {
	InputURL     string
	Result       string
	APIs         int
	RequestOK    int
	RequestSkip  int
	RequestFail  int
	Unauthorized int
	DumpJSON     string
}

func formatScanResultLine(s ScanFinishedSummary) string {
	if strings.TrimSpace(s.Result) != "" {
		return strings.TrimSpace(s.Result)
	}
	line := fmt.Sprintf("API=%d | Request[ok]=%d | Request[skip]=%d | Request[fail]=%d",
		s.APIs, s.RequestOK, s.RequestSkip, s.RequestFail)
	if s.Unauthorized > 0 {
		line += fmt.Sprintf(" | Unauthorized=%d", s.Unauthorized)
	}
	return line
}

func formatDumpJSON(path string) string {
	if strings.TrimSpace(path) == "" {
		return "none"
	}
	return islazy.RelativePathForDisplay(path)
}

// PrintScanFinished logs the summary block (Print-level visibility; direct stderr for layout/color).
func PrintScanFinished(s ScanFinishedSummary) {
	Logger.Helper()
	emitScanFinishedSummary(s)
}

// PrintScanFailed logs resolve/probe failure in the same Scan Finished layout.
func PrintScanFailed(inputURL string, err error) {
	if err == nil {
		return
	}
	PrintScanFinished(ScanFinishedSummary{
		InputURL: inputURL,
		Result:   ScanErrorMessage(err),
	})
}

// ScanErrorMessage unwraps CLI errors into user-facing text (no phase prefixes).
func ScanErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	for {
		changed := false
		for _, p := range []string{
			"resolve swagger url: ",
			"save api-docs: ",
			"analyze swagger: ",
			"http meta: ",
			"write csv: ",
			"write jsonl: ",
			"open database: ",
			"write database: ",
			"scan error: ",
		} {
			if strings.HasPrefix(msg, p) {
				msg = strings.TrimSpace(msg[len(p):])
				changed = true
			}
		}
		if strings.HasPrefix(msg, "[-] ") {
			msg = strings.TrimSpace(msg[4:])
			changed = true
		}
		if !changed {
			break
		}
	}
	return msg
}

// PrintBatchComplete logs batch rollup at Print level.
func PrintBatchComplete(targetsFile string, total, ok, failed int) {
	Logger.Helper()
	emitBatchCompleteSummary(targetsFile, total, ok, failed)
}
