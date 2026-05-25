// single.go: RunSingle single Swagger scan pipeline and result types.
package scanrun

import (
	"encoding/json"
	"fmt"

	"swagger-exp-knife4j/pkg/extension"
	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/scanner"
	"swagger-exp-knife4j/pkg/writers"
)

// WriterConfig specifies where scan results are written; batch scans can enable append mode.
type WriterConfig struct {
	DbURI     string // non-empty writes to this database URI
	CsvFile   string // non-empty writes CSV
	JsonlFile string // non-empty writes JSONL
	DbDebug   bool   // print GORM SQL when true
	AppendCSV   bool // append CSV rows when true and file exists (RunFile from 2nd URL onward)
	AppendJSONL bool // append JSONL lines when true and file exists
}

// SingleParams is input for a single scan.
type SingleParams struct {
	InputURL  string               // Swagger JSON or Knife4j/HTML page URL
	OutputDir string               // root directory for api-docs on disk
	DocsOnly  bool                 // dump OpenAPI JSON only; skip AutoRequestAllAPI and probe writers
	HTTP      *scanner.HTTPOptions // HTTP client config; nil uses concurrency 1
	Writers   WriterConfig         // optional persistence targets
}

// SingleResult summary returned to CLI / MCP tool swagger_scan (JSON text).
// Field meanings: docs/MCP_TOOLS.md and pkg/mcpserver/tools_schema.go.
type SingleResult struct {
	InputURL      string `json:"input_url"`
	ResolvedJSON  string `json:"resolved_json_url"`
	APIDocsPath   string `json:"api_docs_path,omitempty"`
	PathCount     int    `json:"path_count"`
	RequestCount  int    `json:"request_count"`
	RequestOK      int `json:"request_ok"`
	RequestSkipped int `json:"request_skipped"`
	RequestFailed  int `json:"request_failed"`
	Unauthorized   int `json:"unauthorized_count"`
	Stats         []scanner.APIStatisticsInfo `json:"-"`
	WroteDB       bool   `json:"wrote_db"`
	WroteCSV      bool   `json:"wrote_csv"`
	WroteJSONL    bool   `json:"wrote_jsonl"`
	DbURI         string `json:"db_uri,omitempty"`
	CsvFile       string `json:"csv_file,omitempty"`
	JsonlFile     string `json:"jsonl_file,omitempty"`
}

// RunSingle runs the single-scan pipeline (MCP swagger_scan and CLI scan single).
// Order: ResolveSwaggerJSONURL → SaveAPIDocsJSON → AnalyzeSwaggerAPI → AutoRequestAllAPI → optional writes.
// Returns *SingleResult or error; no partial result on success.
func RunSingle(p SingleParams) (*SingleResult, error) {
	if p.InputURL == "" {
		return nil, fmt.Errorf("input URL is required")
	}
	if p.OutputDir == "" {
		p.OutputDir = "output"
	}
	httpOpts := p.HTTP
	if httpOpts == nil {
		httpOpts = &scanner.HTTPOptions{Parallel: 1}
	}
	if err := httpOpts.Validate(); err != nil {
		return nil, err
	}

	scanCtx := &extension.ScanContext{
		InputURL:  p.InputURL,
		OutputDir: p.OutputDir,
		HTTP:      httpOpts,
		DocsOnly:  p.DocsOnly,
	}
	if err := extension.RunScanHooks(extension.PhaseBeforeResolve, scanCtx); err != nil {
		return nil, err
	}

	jsonURL, err := scanner.ResolveSwaggerJSONURL(p.InputURL, httpOpts)
	if err != nil {
		return nil, fmt.Errorf("resolve swagger url: %w", err)
	}
	scanCtx.ResolvedJSONURL = jsonURL
	if err := extension.RunScanHooks(extension.PhaseAfterResolve, scanCtx); err != nil {
		return nil, err
	}

	apiDocsPath, err := scanner.SaveAPIDocsJSON(p.OutputDir, p.InputURL, jsonURL, httpOpts)
	if err != nil {
		return nil, fmt.Errorf("save api-docs: %w", err)
	}
	scanCtx.APIDocsPath = apiDocsPath
	if err := extension.RunScanHooks(extension.PhaseAfterSaveAPIDocs, scanCtx); err != nil {
		return nil, err
	}

	stats, err := scanner.AnalyzeSwaggerAPI(jsonURL, httpOpts)
	if err != nil {
		return nil, fmt.Errorf("analyze swagger: %w", err)
	}
	scanCtx.Stats = stats
	if err := extension.RunScanHooks(extension.PhaseAfterAnalyze, scanCtx); err != nil {
		return nil, err
	}

	var requestResults []scanner.APIRequestResult
	var probe scanner.ProbeSummary
	if p.DocsOnly {
		logScanPipelineDocsOnly(p.InputURL, jsonURL, apiDocsPath, len(stats))
	} else {
		requestResults = scanner.AutoRequestAllAPI(jsonURL, stats, httpOpts)
		probe = scanner.SummarizeProbeResults(requestResults)
		logScanPipeline(p.InputURL, jsonURL, apiDocsPath, len(stats), httpOpts, probe)
	}
	scanCtx.ProbeResults = requestResults
	if err := extension.RunScanHooks(extension.PhaseAfterProbe, scanCtx); err != nil {
		return nil, err
	}

	meta, err := httpOpts.Meta()
	if err != nil {
		return nil, fmt.Errorf("http meta: %w", err)
	}
	scanCtx.HTTPMeta = meta
	if err := extension.RunScanHooks(extension.PhaseBeforeWrite, scanCtx); err != nil {
		return nil, err
	}
	if !p.DocsOnly {
		if err := extension.RunScanWriters(scanCtx); err != nil {
			return nil, err
		}
	} else if p.Writers.DbURI != "" || p.Writers.CsvFile != "" || p.Writers.JsonlFile != "" {
		log.WarnLine("--docs-only skips --write-db/csv/jsonl (no API probe records to persist)")
	}

	out := &SingleResult{
		InputURL:      p.InputURL,
		ResolvedJSON:  jsonURL,
		APIDocsPath:   apiDocsPath,
		PathCount:     len(stats),
		RequestCount:  probe.Total,
		RequestOK:      probe.OK,
		RequestSkipped: probe.Skipped,
		RequestFailed:  probe.Failed,
		Unauthorized:   probe.Unauthorized,
		Stats:         stats,
	}

	if p.Writers.CsvFile != "" && !p.DocsOnly {
		if err := scanner.WriteReportToCSV(p.Writers.CsvFile, p.InputURL, stats, requestResults, p.Writers.AppendCSV); err != nil {
			return nil, fmt.Errorf("write csv: %w", err)
		}
		out.WroteCSV = true
		out.CsvFile = p.Writers.CsvFile
	}

	if p.Writers.JsonlFile != "" && !p.DocsOnly {
		if err := scanner.WriteReportToJSONL(p.Writers.JsonlFile, p.InputURL, jsonURL, meta, stats, requestResults, p.Writers.AppendJSONL); err != nil {
			return nil, fmt.Errorf("write jsonl: %w", err)
		}
		out.WroteJSONL = true
		out.JsonlFile = p.Writers.JsonlFile
	}

	if p.Writers.DbURI != "" && !p.DocsOnly {
		dbWriter, err := writers.NewSwaggerDbWriter(p.Writers.DbURI, p.Writers.DbDebug)
		if err != nil {
			return nil, fmt.Errorf("open database: %w", err)
		}
		records := scanner.BuildSwaggerReportRecords(p.InputURL, jsonURL, meta, stats, requestResults)
		if err := dbWriter.WriteRecords(records); err != nil {
			return nil, fmt.Errorf("write database: %w", err)
		}
		out.WroteDB = true
		out.DbURI = p.Writers.DbURI
	}

	if err := extension.RunScanHooks(extension.PhaseAfterWrite, scanCtx); err != nil {
		return nil, err
	}

	return out, nil
}

// SingleResultJSON formats SingleResult as indented JSON for MCP swagger_scan responses.
// On success returns JSON text; on failure returns error.
func SingleResultJSON(r *SingleResult) (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
