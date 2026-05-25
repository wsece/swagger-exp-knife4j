// swagger_docs_save.go: download and save formatted api-docs.json under the output tree.
package scanner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"swagger-exp-knife4j/internal/islazy"
)

// SaveAPIDocsJSON fetches jsonURL and writes {outputDir}/{host}/{scope}/api-docs.json.
// Returns the on-disk path; same host+scope overwrites, different scopes coexist.
func SaveAPIDocsJSON(outputDir, inputURL, jsonURL string, httpOpts *HTTPOptions) (string, error) {
	client, err := httpOpts.Client(15 * time.Second)
	if err != nil {
		return "", err
	}

	body, _, err := fetchWithOptions(client, httpOpts, jsonURL)
	if err != nil {
		return "", fmt.Errorf("fetch api docs: %w", err)
	}
	if !isSwaggerJSON(body) {
		return "", fmt.Errorf("response is not valid swagger/openapi json")
	}

	body = formatJSONBody(body)

	scannedAt := time.Now()
	dest := islazy.ScanAPIDocsPath(outputDir, scannedAt, jsonURL)
	scopeDir := filepath.Dir(dest)

	host := hostFromJSONURL(jsonURL)
	scope := islazy.APIScopeDir(jsonURL)
	manifest := islazy.DocManifest{
		Host:      host,
		Scope:     scope,
		InputURL:  inputURL,
		JSONURL:   jsonURL,
		ScannedAt: scannedAt,
	}
	if manifestBytes, err := json.MarshalIndent(manifest, "", "  "); err == nil {
		_, _ = islazy.WriteFileWithDir(filepath.Join(scopeDir, islazy.ManifestFileName), manifestBytes)
	}

	path, err := islazy.WriteFileWithDir(dest, body)
	if err != nil {
		return "", err
	}

	return path, nil
}

func formatJSONBody(body []byte) []byte {
	var v interface{}
	if json.Unmarshal(body, &v) != nil {
		return body
	}
	formatted, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return body
	}
	return formatted
}

func hostFromJSONURL(jsonURL string) string {
	u, err := url.Parse(jsonURL)
	if err != nil {
		return ""
	}
	return u.Host
}
