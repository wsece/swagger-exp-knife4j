package scanner

import (
	"os"
	"path/filepath"

	"swagger-exp-knife4j/internal/islazy"
)

// ensureReportParentDir creates parent directories for a report output file (does not touch the file itself).
func ensureReportParentDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}
	_, err := islazy.CreateDir(dir)
	return err
}

// reportFileAppendMode decides whether to append to an existing non-empty report file.
func reportFileAppendMode(filePath string, appendMode bool) (absPath string, appendToFile bool, err error) {
	if err := ensureReportParentDir(filePath); err != nil {
		return "", false, err
	}
	absPath, err = filepath.Abs(filePath)
	if err != nil {
		return filePath, appendMode, nil
	}
	if !appendMode {
		return absPath, false, nil
	}
	st, statErr := os.Stat(absPath)
	if statErr == nil && st.Size() > 0 {
		return absPath, true, nil
	}
	return absPath, false, nil
}
