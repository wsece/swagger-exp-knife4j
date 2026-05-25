// progress_log.go: user-facing scan pipeline steps on stderr.
package scanrun

import (
	"fmt"
	"strings"
	"time"

	"swagger-exp-knife4j/internal/islazy"
	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/scanner"
)

func logScanPipelineDocsOnly(inputURL, jsonURL, apiDocsPath string, apiCount int) {
	if inputURL != "" && !sameSwaggerEntry(inputURL, jsonURL) {
		log.InfoStep("Found SwaggerUI", "page: "+inputURL)
	}
	log.InfoStep("Found OpenAPI", "json: "+jsonURL)
	if apiDocsPath != "" {
		log.InfoStep("Dump OpenAPI", "file: "+islazy.RelativePathForDisplay(apiDocsPath))
	}
	log.InfoStep("Found API number", fmt.Sprintf("Api:%d", apiCount))
	log.InfoStep("Docs only", "skipped automated API requests")
}

func logScanPipeline(inputURL, jsonURL, apiDocsPath string, apiCount int, httpOpts *scanner.HTTPOptions, probe scanner.ProbeSummary) {
	if inputURL != "" && !sameSwaggerEntry(inputURL, jsonURL) {
		log.InfoStep("Found SwaggerUI", "page: "+inputURL)
	}
	log.InfoStep("Found OpenAPI", "json: "+jsonURL)
	if apiDocsPath != "" {
		log.InfoStep("Dump OpenAPI", "file: "+islazy.RelativePathForDisplay(apiDocsPath))
	}
	log.InfoStep("Found API number", fmt.Sprintf("Api:%d", apiCount))

	parallel := 1
	delay := time.Duration(0)
	if httpOpts != nil {
		if httpOpts.Parallel > 0 {
			parallel = httpOpts.Parallel
		}
		delay = httpOpts.Delay
	}
	log.InfoStep("Request automate",
		fmt.Sprintf("swagger: %s | parallel=%d | delay=%s", jsonURL, parallel, delay),
	)

	detail := fmt.Sprintf("total=%d | ok=%d | skip=%d | fail=%d",
		probe.Total, probe.OK, probe.Skipped, probe.Failed)
	if probe.Unauthorized > 0 {
		detail += fmt.Sprintf(" | unauthorized=%d", probe.Unauthorized)
	}
	log.InfoStep("Request finished", detail)
}

func sameSwaggerEntry(page, jsonURL string) bool {
	page = strings.TrimSpace(strings.TrimSuffix(page, "/"))
	jsonURL = strings.TrimSpace(strings.TrimSuffix(jsonURL, "/"))
	return page != "" && strings.EqualFold(page, jsonURL)
}
