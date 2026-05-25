// scan_summary_style.go: terminal styling for Scan Finished / Batch summary blocks.
package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"swagger-exp-knife4j/internal/termcolor"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

var (
	scanTitleOK    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))  // green
	scanTitleFail  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")) // red
	scanTitleBatch = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))  // blue
	scanLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	scanValueStyle = lipgloss.NewStyle()
	scanURLStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // yellow, match warn highlight
	scanPathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	scanMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	scanResultOK   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	scanResultWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	scanResultFail = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

func summaryTimestamp() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

func summaryColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isatty.IsTerminal(os.Stderr.Fd())
}

// renderStyled applies lipgloss even when termcolor.Enabled is false (Windows tty),
// because global lipgloss profile may be Ascii while charm log levels still colorize.
func renderStyled(style lipgloss.Style, s string) string {
	if s == "" || !summaryColorEnabled() {
		return s
	}
	saved := lipgloss.ColorProfile()
	if termcolor.Enabled {
		lipgloss.SetColorProfile(termenv.TrueColor)
	} else {
		lipgloss.SetColorProfile(termenv.ANSI256)
	}
	out := style.Render(s)
	lipgloss.SetColorProfile(saved)
	return out
}

func isScanFailureSummary(s ScanFinishedSummary) bool {
	return strings.TrimSpace(s.Result) != ""
}

func renderSummaryRow(label, value string, valueStyle lipgloss.Style) string {
	return "\t" + renderStyled(scanLabelStyle, label+":") + " " + renderStyled(valueStyle, value)
}

func emitSummaryBlock(title string, titleStyle lipgloss.Style, detailRows []string) {
	_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", summaryTimestamp(), renderStyled(titleStyle, title))
	for _, row := range detailRows {
		_, _ = fmt.Fprintln(os.Stderr, row)
	}
}

func emitScanFinishedSummary(s ScanFinishedSummary) {
	titleStyle := scanTitleOK
	if isScanFailureSummary(s) {
		titleStyle = scanTitleFail
	}

	resultText := formatScanResultLine(s)
	resultStyle := scanResultOK
	switch {
	case isScanFailureSummary(s):
		resultStyle = scanResultFail
	case s.Unauthorized > 0:
		resultStyle = scanResultWarn
	}

	dump := formatDumpJSON(s.DumpJSON)
	dumpStyle := scanPathStyle
	if dump == "none" {
		dumpStyle = scanMutedStyle
	}

	emitSummaryBlock("Scan Finished", titleStyle, []string{
		renderSummaryRow("Input", s.InputURL, scanURLStyle),
		renderSummaryRow("Result", resultText, resultStyle),
		renderSummaryRow("DumpJson", dump, dumpStyle),
	})
}

func emitBatchCompleteSummary(targetsFile string, total, ok, failed int) {
	resultStyle := scanResultOK
	if failed > 0 && ok == 0 {
		resultStyle = scanResultFail
	} else if failed > 0 {
		resultStyle = scanResultWarn
	}
	emitSummaryBlock("Batch Scan Finished", scanTitleBatch, []string{
		renderSummaryRow("Targets", targetsFile, scanURLStyle),
		renderSummaryRow("Result", fmt.Sprintf("total=%d | ok=%d | failed=%d", total, ok, failed), resultStyle),
	})
}
