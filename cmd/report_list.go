// report_list.go: Implements the `report list` subcommand.
// Reads records from the database and displays them in a terminal table format.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"swagger-exp-knife4j/internal/islazy"
	"swagger-exp-knife4j/internal/termcolor"
	"swagger-exp-knife4j/pkg/database"
	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/models"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var reportListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API scan records from the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReportList(reportOpts.DbURI)
	},
}

func init() {
	SetCommandHelp(reportListCmd, helpReportListLong, helpReportListExample)
	reportCmd.AddCommand(reportListCmd)
}

// runReportList connects to the database using dbURI, queries records sorted by scanned_at descending, and renders the terminal table.
func runReportList(dbURI string) error {
	conn, err := database.SwaggerConnection(dbURI, false)
	if err != nil {
		return err
	}

	var records []models.SwaggerAPIRecord
	if err := conn.Order("scanned_at desc, id desc").Find(&records).Error; err != nil {
		return err
	}

	if len(records) == 0 {
		log.Info("no records in database", "uri", dbURI)
		return nil
	}

	log.Info("listing swagger api records", "uri", dbURI, "count", len(records))
	fmt.Fprintln(os.Stdout, renderRecordTable(records))
	return nil
}

func renderRecordTable(records []models.SwaggerAPIRecord) string {
	headers := []string{"WHEN", "INPUT URL", "METHOD", "HOST", "API", "STATUS"}
	widths := []int{16, 28, 8, 22, 40, 8}

	rows := make([][]string, 0, len(records))
	for _, r := range records {
		rows = append(rows, []string{
			r.ScannedAt.Format("2006-01-02 15:04"),
			islazy.Truncate(r.InputURL, widths[1]),
			r.Method,
			r.Host,
			r.API,
			formatListStatus(r),
		})
	}

	return drawRoundedTable(headers, rows, widths)
}

func formatListStatus(r models.SwaggerAPIRecord) string {
	if r.Failed && r.StatusCode == 0 {
		return failedStatus()
	}
	return statusCode(r.StatusCode)
}

func statusCode(code int) string {
	s := fmt.Sprintf("%d", code)
	if !termcolor.Enabled {
		return s
	}
	var style lipgloss.Style
	switch {
	case code >= 200 && code < 300:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case code >= 300 && code < 400:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	case code >= 400 && code < 500:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	case code >= 500:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	}
	return style.Render(s)
}

func failedStatus() string {
	if !termcolor.Enabled {
		return "FAIL"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("FAIL")
}

func drawRoundedTable(headers []string, rows [][]string, widths []int) string {
	padCell := func(s string, w int) string {
		gap := w - lipgloss.Width(s)
		if gap > 0 {
			return s + strings.Repeat(" ", gap)
		}
		return s
	}

	var lines []string
	headerCells := make([]string, len(headers))
	for i, h := range headers {
		cell := padCell(h, widths[i])
		if termcolor.Enabled {
			cell = lipgloss.NewStyle().Bold(true).Render(cell)
		}
		headerCells[i] = cell
	}
	lines = append(lines, strings.Join(headerCells, "  "))

	sep := strings.Repeat("─", sum(widths)+2*(len(widths)-1))
	lines = append(lines, sep)

	for _, row := range rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = padCell(cell, widths[i])
		}
		lines = append(lines, strings.Join(cells, "  "))
	}

	body := strings.Join(lines, "\n")
	if !termcolor.Enabled {
		return body
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(body)
}

func sum(nums []int) int {
	n := 0
	for _, v := range nums {
		n += v
	}
	return n
}
