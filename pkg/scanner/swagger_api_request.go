// swagger_api_request.go: concurrent auto GET/POST requests, CSV export, and APIRequestResult.
package scanner

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"swagger-exp-knife4j/pkg/log"
)

// APIRequestResult is the outcome of one automated request
type APIRequestResult struct {
	Host            string
	Method          string
	Path            string
	FullURL         string
	FinalURL        string
	RequestParams   string
	RequestBody     string
	RequestHeaders  []HTTPHeaderKV
	ResponseHeaders []HTTPHeaderKV
	StatusCode      int
	ContentType     string
	Response        string
	DurationMs      int64
	Error           string
}

// AutoRequestAllAPI concurrently requests each GET/POST in apiList per httpOpts.Parallel.
// swaggerURL is for logging; httpOpts controls headers, delay, timeout, and concurrency.
// Returns one APIRequestResult per API (includes RequestHeaders/ResponseHeaders for capture).
func AutoRequestAllAPI(swaggerURL string, apiList []APIStatisticsInfo, httpOpts *HTTPOptions) []APIRequestResult {
	if httpOpts != nil {
		if err := httpOpts.Validate(); err != nil {
			log.Error("invalid http options", "err", err)
			return nil
		}
	}

	u, err := url.Parse(swaggerURL)
	if err != nil {
		log.Error("[-] Found Swagger URL err:", "url", swaggerURL, "err", err)
		return nil
	}
	baseHost := u.Scheme + "://" + u.Host

	// API probe: when -m is unset, no overall request timeout (fallback 0)
	client, err := httpOpts.Client(0)
	if err != nil {
		log.Error("[-] Create HTTP Server err :", "err", err)
		return nil
	}

	results := make([]APIRequestResult, len(apiList))
	jobs := make(chan int, len(apiList))
	for i := range apiList {
		jobs <- i
	}
	close(jobs)

	workers := parallelWorkers(httpOpts)
	delay := requestDelay(httpOpts)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = executeAPIRequest(client, httpOpts, u.Host, baseHost, apiList[i])
				if delay > 0 {
					time.Sleep(delay)
				}
			}
		}()
	}
	wg.Wait()

	logUnauthorizedCandidates(results)
	return results
}

// logUnauthorizedCandidates logs each non-401 response at Debug (summary count is on scanrun.SingleResult).
func logUnauthorizedCandidates(results []APIRequestResult) {
	for _, r := range results {
		if r.StatusCode == 0 || r.StatusCode == http.StatusUnauthorized {
			continue
		}
		host := r.Host
		if host == "" {
			host = hostFromURL(r.FullURL)
		}
		log.Debug("unauthorized candidate", "host", host, "path", r.Path, "status", r.StatusCode)
	}
}

func hostFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	return u.Host
}

// WriteReportToCSV writes one row per API; first column is the input Swagger URL.
// When appendMode is true and the file already exists, appends data rows (no header).
func WriteReportToCSV(filePath, inputURL string, stats []APIStatisticsInfo, requests []APIRequestResult, appendMode bool) error {
	absPath, appendToFile, err := reportFileAppendMode(filePath, appendMode)
	if err != nil {
		return err
	}
	writeHeader := !appendToFile
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

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if writeHeader {
		if err := writer.Write([]string{
			"Swagger URL",
			"Method",
			"API",
			"Parameter  ",
			"URL",
			"Request parameters or request body",
			"Status Code",
			"Response Type",
			"Response Body",
			"Error Message",
		}); err != nil {
			return err
		}
	}

	reqByKey := make(map[string]APIRequestResult, len(requests))
	for _, r := range requests {
		reqByKey[r.Method+"|"+r.Path] = r
	}

	for i, api := range stats {
		req := requestForAPI(api, i, requests, reqByKey)

		status := ""
		if req.StatusCode != 0 {
			status = strconv.Itoa(req.StatusCode)
		}

		if err := writer.Write([]string{
			inputURL,
			api.Method,
			api.Path,
			formatAPIParams(api.Parameters),
			req.FullURL,
			requestPayload(req),
			status,
			req.ContentType,
			req.Response,
			req.Error,
		}); err != nil {
			return err
		}
	}

	return nil
}

func requestPayload(req APIRequestResult) string {
	if req.RequestBody != "" {
		return req.RequestBody
	}
	return req.RequestParams
}

func requestForAPI(api APIStatisticsInfo, index int, requests []APIRequestResult, byKey map[string]APIRequestResult) APIRequestResult {
	key := api.Method + "|" + api.Path
	if r, ok := byKey[key]; ok {
		return r
	}
	if index < len(requests) && requests[index].Method == api.Method && requests[index].Path == api.Path {
		return requests[index]
	}
	return APIRequestResult{Method: api.Method, Path: api.Path}
}

func paramNames(params []APIAutoFillParam) string {
	if len(params) == 0 {
		return ""
	}
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}

func formatAPIParams(params []APIAutoFillParam) string {
	if len(params) == 0 {
		return "No parameters"
	}
	parts := make([]string, 0, len(params))
	for _, p := range params {
		required := "Optional"
		if p.Required {
			required = "Required"
		}
		parts = append(parts, fmt.Sprintf("%s(%s,%s,%s)", p.Name, p.Position, p.Type, required))
	}
	return strings.Join(parts, "; ")
}

// getTestValue returns a test value for the given parameter type
func getTestValue(paramType string) interface{} {
	switch paramType {
	case "string":
		return "test"
	case "int", "integer":
		return 1
	case "bool", "boolean":
		return true
	case "float":
		return 1.0
	default:
		return "test"
	}
}
