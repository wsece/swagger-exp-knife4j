// file.go: RunFile batch scan from a targets file and batch result types.
package scanrun

import (
	"fmt"

	"swagger-exp-knife4j/internal/islazy"
)

// FileResult summarizes a batch scan (CLI scan file and potential MCP extensions).
type FileResult struct {
	Total    int             `json:"total"`
	OK       int             `json:"ok"`
	Failed   int             `json:"failed"`
	Results  []SingleResult  `json:"results"`
	Errors   []FileTargetErr `json:"errors,omitempty"`
}

// FileTargetErr records failure for one URL in a batch scan.
type FileTargetErr struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

// BatchHooks optional callbacks for multi-target scans (CLI progress); nil is safe.
type BatchHooks struct {
	OnTargetStart func(index, total int, url string)
	OnTargetDone  func(index, total int, url string, result *SingleResult, err error)
}

// RunFile reads targetsFile (one URL per line) and calls RunURLs for each target.
func RunFile(targetsFile string, base SingleParams, hooks *BatchHooks) (*FileResult, error) {
	urls, err := islazy.ReadTargetURLs(targetsFile)
	if err != nil {
		return nil, err
	}
	return RunURLs(urls, base, hooks)
}

// RunURLs calls RunSingle for each URL (Web batch dispatch and CLI scan file).
// If all fail, returns *FileResult and error; partial success returns *FileResult only.
func RunURLs(urls []string, base SingleParams, hooks *BatchHooks) (*FileResult, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no targets provided")
	}

	out := &FileResult{Total: len(urls)}
	csvPath := base.Writers.CsvFile
	jsonlPath := base.Writers.JsonlFile

	for i, url := range urls {
		index := i + 1
		total := len(urls)
		if hooks != nil && hooks.OnTargetStart != nil {
			hooks.OnTargetStart(index, total, url)
		}

		p := base
		p.InputURL = url
		// From the 2nd URL onward, append to the same CSV/JSONL (see scanner.WriteReportTo*).
		if csvPath != "" {
			p.Writers.AppendCSV = i > 0
		}
		if jsonlPath != "" {
			p.Writers.AppendJSONL = i > 0
		}

		result, err := RunSingle(p)
		if err != nil {
			out.Failed++
			out.Errors = append(out.Errors, FileTargetErr{URL: url, Error: err.Error()})
			if hooks != nil && hooks.OnTargetDone != nil {
				hooks.OnTargetDone(index, total, url, nil, err)
			}
			continue
		}
		out.OK++
		out.Results = append(out.Results, *result)
		if hooks != nil && hooks.OnTargetDone != nil {
			hooks.OnTargetDone(index, total, url, result, nil)
		}
	}

	if out.OK == 0 {
		return out, fmt.Errorf("all %d targets failed", out.Total)
	}
	return out, nil
}
