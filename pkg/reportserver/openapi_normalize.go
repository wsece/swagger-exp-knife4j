// openapi_normalize.go: normalizes OpenAPI JSON for report preview (servers, host, etc.).
package reportserver

import (
	"encoding/json"
	"strings"
)

// Knife4jProxyServerURL is the OpenAPI servers[0].url for in-browser debug (same-origin proxy).
func Knife4jProxyServerURL(sessionID string) string {
	return "/knife4j/" + sessionID + "/try"
}

// NormalizeOpenAPISpec patches common issues that break Knife4j rendering.
// serverURL is a full origin (https://host) or bare hostname (https:// is added).
func NormalizeOpenAPISpec(data []byte, serverURL string) []byte {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return data
	}

	if serverURL != "" {
		url := serverURL
		if strings.HasPrefix(url, "/") {
			// Same-origin Knife4j proxy base (e.g. /knife4j/{session}/try)
			doc["servers"] = []map[string]string{{"url": url}}
		} else {
			if !strings.Contains(url, "://") {
				url = "https://" + url
			}
			doc["servers"] = []map[string]string{{"url": url}}
		}
	}

	fixOpenAPINode(doc)

	out, err := json.Marshal(doc)
	if err != nil {
		return data
	}
	return out
}

func fixOpenAPINode(v interface{}) {
	switch node := v.(type) {
	case map[string]interface{}:
		ensureSchemaType(node)
		for _, child := range node {
			fixOpenAPINode(child)
		}
	case []interface{}:
		for _, item := range node {
			fixOpenAPINode(item)
		}
	}
}

func ensureSchemaType(m map[string]interface{}) {
	if _, ok := m["$ref"]; ok {
		return
	}
	if _, ok := m["type"]; ok {
		return
	}
	if _, ok := m["properties"]; ok {
		m["type"] = "object"
		return
	}
	if _, ok := m["items"]; ok {
		m["type"] = "array"
		return
	}
	if _, ok := m["allOf"]; ok {
		m["type"] = "object"
		return
	}
	if _, ok := m["anyOf"]; ok {
		return
	}
	if _, ok := m["oneOf"]; ok {
		return
	}
	if _, ok := m["enum"]; ok {
		m["type"] = "string"
		return
	}
	if _, ok := m["additionalProperties"]; ok {
		m["type"] = "object"
		return
	}
	if _, ok := m["nullable"]; ok {
		m["type"] = "object"
	}
}
