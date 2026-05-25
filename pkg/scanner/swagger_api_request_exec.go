// swagger_api_request_exec.go: per-API HTTP execution (executeAPIRequest) and concurrency/delay helpers.
package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"swagger-exp-knife4j/pkg/log"
)

func executeAPIRequest(client *http.Client, httpOpts *HTTPOptions, host, baseHost string, api APIStatisticsInfo) APIRequestResult {
	fullURL := baseHost + api.Path
	method := strings.ToUpper(api.Method)

	row := APIRequestResult{
		Host:    host,
		Method:  method,
		Path:    api.Path,
		FullURL: fullURL,
	}

	reqLogger := log.With("method", method, "path", api.Path, "url", fullURL)
	reqLogger.Debug("sending request")

	var req *http.Request

	switch method {
	case "GET":
		req, _ = http.NewRequest(http.MethodGet, fullURL, nil)
		q := req.URL.Query()
		for _, param := range api.Parameters {
			if param.Position == "query" {
				q.Add(param.Name, fmt.Sprintf("%v", getTestValue(param.Type)))
			}
		}
		req.URL.RawQuery = q.Encode()
		row.RequestParams = req.URL.RawQuery
		row.FinalURL = req.URL.String()
		reqLogger.Debug("request query", "query", row.RequestParams)

	case "POST":
		bodyMap := make(map[string]interface{})
		for _, param := range api.Parameters {
			if param.Position == "body" {
				bodyMap[param.Name] = getTestValue(param.Type)
			}
		}
		jsonBuf, _ := json.Marshal(bodyMap)
		row.RequestBody = string(jsonBuf)
		reqLogger.Debug("request body", "body", row.RequestBody)

		var bodyErr error
		req, bodyErr = http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(jsonBuf))
		if bodyErr != nil {
			row.Error = bodyErr.Error()
			reqLogger.Error("build request failed", "err", bodyErr)
			return row
		}
	default:
		row.Error = "Skip non-GET/POST methods"
		reqLogger.Warn("skipped", "reason", row.Error)
		return row
	}

	if err := httpOpts.Apply(req); err != nil {
		row.Error = err.Error()
		reqLogger.Error("apply request options failed", "err", err)
		return row
	}
	if method == http.MethodPost && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	row.FinalURL = req.URL.String()
	row.RequestHeaders = headersToKV(req.Header)

	start := time.Now()
	resp, err := client.Do(req)
	row.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		row.Error = err.Error()
		reqLogger.Error("request failed", "err", err)
		return row
	}

	row.StatusCode = resp.StatusCode
	row.ContentType = resp.Header.Get("Content-Type")
	row.ResponseHeaders = headersToKV(resp.Header)

	respBodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var respBody interface{}
	if json.Unmarshal(respBodyBytes, &respBody) == nil {
		prettyResp, _ := json.Marshal(respBody)
		row.Response = string(prettyResp)
	} else {
		row.Response = string(respBodyBytes)
	}

	reqLogger.Debug("response received",
		"status", row.StatusCode,
		"content-type", row.ContentType,
	)
	reqLogger.Debug("response body", "body", row.Response)

	return row
}

func parallelWorkers(httpOpts *HTTPOptions) int {
	if httpOpts == nil || httpOpts.Parallel < 1 {
		return 1
	}
	return httpOpts.Parallel
}

func requestDelay(httpOpts *HTTPOptions) time.Duration {
	if httpOpts == nil || httpOpts.Delay <= 0 {
		return 0
	}
	return httpOpts.Delay
}
