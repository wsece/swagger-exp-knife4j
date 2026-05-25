// Package log wraps charmbracelet/log for CLI and server output.
//
// Level guide (use consistently across the repo):
//
//   - Print: user-facing outcome that must appear even with --quiet (-q).
//     Examples: scan complete summary, unauthorized-access hit count, report server URL.
//
//   - Info: normal progress hidden by --quiet.
//     Examples: scan pipeline steps (InfoStep), writer paths, batch target markers.
//
//   - Debug: verbose detail; enable with --debug-log (-D).
//     Examples: per-API parameters, probe candidates, batch per-target lines.
//
//   - Warn / Error: non-fatal issues and operational failures (see call sites).
//
// Command primary data (tables, JSON) should go to stdout (cmd.OutOrStdout), not log.
package log

import (
	"os"

	"swagger-exp-knife4j/internal/termcolor"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// LLogger is a type alias for charmbracelet/log.Logger.
type LLogger = log.Logger

// Logger is the package-global logger, created in init.
var Logger *LLogger

func init() {
	styles := log.DefaultStyles()
	if termcolor.Enabled {
		styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
		styles.Values["err"] = lipgloss.NewStyle().Bold(true)
	}

	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
	})
	Logger.SetStyles(styles)
	Logger.SetLevel(log.InfoLevel)
}

// EnableDebug enables Debug level and logs caller.
func EnableDebug() {
	Logger.SetLevel(log.DebugLevel)
	Logger.SetReportCaller(true)
}

// EnableSilence suppresses Info/Warn/Debug; Print still outputs.
func EnableSilence() {
	Logger.SetLevel(log.FatalLevel + 100)
}

// Debug logs debug messages
func Debug(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Debug(msg, keyvals...)
}

// Info logs info messages
func Info(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Info(msg, keyvals...)
}

// Warn logs warning messages
func Warn(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Warn(msg, keyvals...)
}

// Error logs error messages
func Error(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Error(msg, keyvals...)
}

// Fatal logs fatal messages and panics
func Fatal(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Fatal(msg, keyvals...)
}

// Print logs messages regardless of level (--quiet does not suppress).
func Print(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Print(msg, keyvals...)
}

// With returns a sublogger with a prefix
func With(keyvals ...interface{}) *LLogger {
	return Logger.With(keyvals...)
}
