// Package scanner implements core Swagger/Knife4j scanning (library layer).
//
// Modules (see file header comments in this package):
//
//   - swagger_url_resolve: resolve OpenAPI JSON URL from HTML or direct URL
//   - swagger_docs_save: save spec JSON under output/{host}/{scope}/api-docs.json
//   - swagger_api_statistics: parse paths and parameters, build API test list
//   - swagger_api_request / swagger_api_request_exec: concurrent probe requests and results
//   - swagger_report_jsonl: CSV/JSONL rows and SwaggerReportRecord
//   - http_exchange: Burp-style meta+request+response capture struct
//   - http_options: curl-style HTTP client config (headers, cookies, proxy, timeout, concurrency)
package scanner
