// swagger_url_resolve.go: resolve OpenAPI document URL from Swagger UI/Knife4j HTML or direct JSON.
package scanner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"swagger-exp-knife4j/pkg/log"
)

const maxSwaggerFetchBytes = 8 << 20

var swaggerURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\burl\s*:\s*["']([^"']+)["']`),
	regexp.MustCompile(`(?i)\bconfigUrl\s*:\s*["']([^"']+)["']`),
	regexp.MustCompile(`(?i)\bswaggerUrl\s*:\s*["']([^"']+)["']`),
	regexp.MustCompile(`(?i)\bswagger-ui-url\s*=\s*["']([^"']+)["']`),
	regexp.MustCompile(`(?i)["']([^"']*/(?:v3/api-docs|v2/api-docs|swagger\.json|openapi\.json)(?:\?[^"']*)?)["']`),
	regexp.MustCompile(`(?i)(?:href|src|data-url)\s*=\s*["']([^"']+\.json[^"']*)["']`),
}

// ResolveSwaggerJSONURL resolves a fetchable OpenAPI/Swagger JSON URL.
// inputURL may be direct JSON or an HTML page; httpOpts fetches pages and probes candidate URLs.
// Returns the final JSON URL or an error on resolve/HTTP failure.
func ResolveSwaggerJSONURL(inputURL string, httpOpts *HTTPOptions) (string, error) {
	inputURL = strings.TrimSpace(inputURL)
	if inputURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	pageURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("Invalid URL: %w", err)
	}

	client, err := httpOpts.Client(15 * time.Second)
	if err != nil {
		return "", err
	}
	body, contentType, err := fetchWithOptions(client, httpOpts, inputURL)
	if err != nil {
		return "", err
	}

	if isSwaggerJSON(body) {
		return inputURL, nil
	}

	if !isHTMLContent(contentType, body) {
		return "", fmt.Errorf("[-] The URL is neither a Swagger UI page nor a Swagger JSON address.")
	}

	candidates := extractSwaggerJSONCandidates(string(body), pageURL)
	for _, candidate := range candidates {
		jsonURL := resolveReference(pageURL, candidate)
		if jsonURL == "" || jsonURL == inputURL {
			continue
		}
		doc, _, err := fetchWithOptions(client, httpOpts, jsonURL)
		if err != nil {
			log.Debug("candidate not reachable", "url", jsonURL, "err", err)
			continue
		}
		if isSwaggerJSON(doc) {
			return jsonURL, nil
		}
	}

	return "", fmt.Errorf("[-] Can't found the Swagger JSON URL. Please check the input URL or try to specify the JSON URL directly.")
}

func isSwaggerJSON(body []byte) bool {
	body = bytesTrimSpace(body)
	if len(body) == 0 || body[0] != '{' {
		return false
	}
	var doc map[string]json.RawMessage
	if json.Unmarshal(body, &doc) != nil {
		return false
	}
	_, hasPaths := doc["paths"]
	return hasPaths
}

func isHTMLContent(contentType string, body []byte) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml") {
		return true
	}
	snippet := strings.ToLower(string(bytesTrimSpace(body)))
	return strings.HasPrefix(snippet, "<!doctype") || strings.HasPrefix(snippet, "<html")
}

func extractSwaggerJSONCandidates(html string, pageURL *url.URL) []string {
	seen := make(map[string]struct{})
	var out []string

	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "javascript:") {
			return
		}
		if _, ok := seen[raw]; ok {
			return
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}

	for _, re := range swaggerURLPatterns {
		for _, m := range re.FindAllStringSubmatch(html, -1) {
			if len(m) > 1 {
				add(m[1])
			}
		}
	}

	for _, p := range commonSwaggerJSONPaths(pageURL) {
		add(p)
	}

	return out
}

func commonSwaggerJSONPaths(pageURL *url.URL) []string {
	base := pageURL.Scheme + "://" + pageURL.Host
	dir := path.Dir(pageURL.Path)
	if dir == "." {
		dir = "/"
	}

	paths := []string{
		"/v3/api-docs",
		"/v2/api-docs",
		"/swagger.json",
		"/api-docs",
		"/openapi.json",
	}

	if dir != "/" {
		paths = append(paths,
			path.Join(dir, "swagger.json"),
			path.Join(dir, "v3/api-docs"),
			path.Join(dir, "v2/api-docs"),
			path.Join(strings.TrimSuffix(dir, "/index.html"), "swagger.json"),
			path.Join(strings.TrimSuffix(dir, "/doc.html"), "v2/api-docs"),
		)
	}

	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = append(out, base+p)
	}
	return out
}

func resolveReference(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	return base.ResolveReference(u).String()
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
