// termcolor.go: Detects terminal ANSI support and configures lipgloss/termenv
// (no color by default on Windows unless FORCE_COLOR=1 is set).
//
// Package termcolor provides disableable colored output for CLI tables and logs.
package termcolor

import (
	"os"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

// Enabled is false when output should be plain text (no ANSI escapes).
var Enabled bool

func init() {
	ConfigureOutput()
}

// ConfigureOutput re-detects color support (call before printing help to stdout).
func ConfigureOutput() {
	Enabled = detectColor(os.Stdout) && detectColor(os.Stderr)
	if !Enabled {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

func detectColor(out *os.File) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
		return true
	}
	if out == nil || !isatty.IsTerminal(out.Fd()) {
		return false
	}
	// Windows classic conhost does not render 256-color glamour output reliably.
	// Opt in only with FORCE_COLOR=1 (see above).
	if runtime.GOOS == "windows" {
		return false
	}
	return true
}
