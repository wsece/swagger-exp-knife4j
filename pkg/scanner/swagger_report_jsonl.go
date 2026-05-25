// swagger_report_jsonl.go: SwaggerReportRecord and shared JSONL/DB report row build/write helpers.
package scanner

import (
	"encoding/json"
	"os"
	"strings"
)

// SwaggerReportRecord is one report row (shared by CSV / JSONL / DB)
type SwaggerReportRecord struct {
	SwaggerURL        string             `json:"swagger_url"`
	ResolvedJSONURL   string             `json:"resolved_json_url,omitempty"`
	Method            string             `json:"method"`
	Path              string             `json:"path"`
	Host              string             `json:"host,omitempty"`
	Parameters        []APIAutoFillParam `json:"parameters,omitempty"`
	ParametersSummary string             `json:"parameters_summary"`
	ParameterNames    string             `json:"parameter_names"`
	Cookie            string             `json:"cookie,omitempty"`
	UserAgent         string             `json:"user_agent,omitempty"`
	Authorization     string             `json:"authorization,omitempty"`
	RequestBody       string             `json:"request_body,omitempty"`
	RequestParams     string             `json:"request_params,omitempty"`
	FullURL           string             `json:"full_url,omitempty"`
	StatusCode        int                `json:"status_code,omitempty"`
	ContentType       string             `json:"content_type,omitempty"`
	Response          string             `json:"response,omitempty"`
	Error             string             `json:"error,omitempty"`
	Failed            bool               `json:"failed"`
	FailedReason      string             `json:"failed_reason,omitempty"`
	UnauthorizedRisk  bool               `json:"unauthorized_risk,omitempty"`
	Exchange          HTTPExchange       `json:"exchange,omitempty"`
}

// WriteReportToJSONL writes scan results line-by-line to JSONL.
// When appendMode is true and the file already exists, appends at EOF.
func WriteReportToJSONL(filePath, inputURL, jsonURL string, meta HTTPRequestMeta, stats []APIStatisticsInfo, requests []APIRequestResult, appendMode bool) error {
	absPath, appendToFile, err := reportFileAppendMode(filePath, appendMode)
	if err != nil {
		return err
	}
	flags := os.O_CREATE | os.O_WRONLY
	if appendToFile {
		flags = os.O_APPEND | os.O_WRONLY
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(absPath, flags, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	records := BuildSwaggerReportRecords(inputURL, jsonURL, meta, stats, requests)
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(line, '\n')); err != nil {
			return err
		}
	}

	return nil
}

// BuildSwaggerReportRecords merges stats and request results into []SwaggerReportRecord (with Exchange capture).
// inputURL is user input; jsonURL is the resolved spec URL; meta is Cookie/UA/Authorization.
func BuildSwaggerReportRecords(inputURL, jsonURL string, meta HTTPRequestMeta, stats []APIStatisticsInfo, requests []APIRequestResult) []SwaggerReportRecord {
	reqByKey := make(map[string]APIRequestResult, len(requests))
	for _, r := range requests {
		reqByKey[r.Method+"|"+r.Path] = r
	}

	records := make([]SwaggerReportRecord, 0, len(stats))
	for i, api := range stats {
		req := requestForAPI(api, i, requests, reqByKey)
		failed := strings.TrimSpace(req.Error) != ""

		host := req.Host
		if host == "" {
			host = hostFromURL(req.FullURL)
		}

		rec := SwaggerReportRecord{
			SwaggerURL:        inputURL,
			ResolvedJSONURL:   jsonURL,
			Method:            api.Method,
			Path:              api.Path,
			Host:              host,
			Parameters:        api.Parameters,
			ParametersSummary: formatAPIParams(api.Parameters),
			ParameterNames:    paramNames(api.Parameters),
			Cookie:            meta.Cookie,
			UserAgent:         meta.UserAgent,
			Authorization:     meta.Authorization,
			RequestBody:       req.RequestBody,
			RequestParams:     req.RequestParams,
			FullURL:           req.FullURL,
			StatusCode:        req.StatusCode,
			ContentType:       req.ContentType,
			Response:          req.Response,
			Error:             req.Error,
			Failed:            failed,
			FailedReason:      req.Error,
			UnauthorizedRisk:  !failed && req.StatusCode != 0 && req.StatusCode != 401,
		}
		rec.Exchange = BuildHTTPExchange(rec, req)
		records = append(records, rec)
	}
	return records
}
