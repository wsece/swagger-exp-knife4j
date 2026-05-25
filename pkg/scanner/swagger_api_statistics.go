// swagger_api_statistics.go: parse OpenAPI paths, resolve refs, build API test list, print summary.
package scanner

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"swagger-exp-knife4j/pkg/log"
)

// SwaggerAPIFile supports OpenAPI 3 fully
type SwaggerAPIFile struct {
	Paths      map[string]map[string]*APIPathItem `json:"paths"`
	Components struct {
		Schemas map[string]*APISchema `json:"schemas"`
	} `json:"components"`
}

// APIPathItem is one HTTP method entry under OpenAPI paths.
type APIPathItem struct {
	Parameters  []APIParameter  `json:"parameters"`
	RequestBody *APIRequestBody `json:"requestBody"`
}

type APIParameter struct {
	Name     string     `json:"name"`
	In       string     `json:"in"`
	Required bool       `json:"required"`
	Type     string     `json:"type"`
	Schema   *APISchema `json:"schema"`
}

type APIRequestBody struct {
	Content map[string]*APIContentType `json:"content"`
}

type APIContentType struct {
	Schema *APISchema `json:"schema"`
}

type APISchema struct {
	Type       string                `json:"type"`
	Properties map[string]*APISchema `json:"properties"`
	Required   []string              `json:"required"`
	Items      *APISchema            `json:"items"`
	Ref        string                `json:"$ref"`
}

// APIStatisticsInfo is one API to test: method, path, and parameters for auto-fill.
type APIStatisticsInfo struct {
	Method     string             `json:"method"`
	Path       string             `json:"path"`
	Parameters []APIAutoFillParam `json:"parameters"`
}

// APIAutoFillParam is one parameter filled during auto-probe (name, in, type, required).
type APIAutoFillParam struct {
	Name     string `json:"name"`
	Position string `json:"position"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// AnalyzeSwaggerAPI downloads and parses the OpenAPI doc at swaggerURL; returns deduplicated APIs.
func AnalyzeSwaggerAPI(swaggerURL string, httpOpts *HTTPOptions) ([]APIStatisticsInfo, error) {
	client, err := httpOpts.Client(15 * time.Second)
	if err != nil {
		return nil, err
	}

	body, _, err := fetchWithOptions(client, httpOpts, swaggerURL)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %w", err)
	}

	var doc SwaggerAPIFile
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("Parse failed: %w", err)
	}

	var stats []APIStatisticsInfo

	for path, methods := range doc.Paths {
		for method, item := range methods {
			info := APIStatisticsInfo{
				Method: strings.ToUpper(method),
				Path:   path,
			}

			// parse query/path parameters
			for _, p := range item.Parameters {
				info.Parameters = append(info.Parameters, APIAutoFillParam{
					Name:     p.Name,
					Position: p.In,
					Type:     getParamType(p.Schema, &doc),
					Required: p.Required,
				})
			}

			// parse body (cycle-safe)
			if item.RequestBody != nil {
				for _, c := range item.RequestBody.Content {
					if c.Schema == nil {
						continue
					}
					schema := resolveRef(c.Schema, &doc)
					// cycle guard: track visited schemas in a map
					visited := make(map[*APISchema]bool)
					parseBodyParams("", schema, &doc, visited, &info.Parameters)
				}
			}

			// dedupe parameters so each appears once
			info.Parameters = uniqueParams(info.Parameters)
			stats = append(stats, info)
		}
	}

	return stats, nil
}

// resolveRef resolves a $ref
func resolveRef(s *APISchema, doc *SwaggerAPIFile) *APISchema {
	if s == nil || doc == nil {
		return s
	}
	if s.Ref != "" {
		parts := strings.Split(s.Ref, "/")
		name := parts[len(parts)-1]
		if sch, ok := doc.Components.Schemas[name]; ok {
			return sch
		}
	}
	return s
}

// getParamType returns the parameter type
func getParamType(s *APISchema, doc *SwaggerAPIFile) string {
	s = resolveRef(s, doc)
	if s == nil {
		return "unknown"
	}

	switch s.Type {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "unknown"
	}
}

// parseBodyParams recursively parses body parameters (cycle-safe)
func parseBodyParams(parent string, s *APISchema, doc *SwaggerAPIFile, visited map[*APISchema]bool, params *[]APIAutoFillParam) {
	s = resolveRef(s, doc)
	if s == nil || s.Properties == nil {
		return
	}

	// cycle guard: skip if this schema was already visited
	if visited[s] {
		return
	}
	visited[s] = true

	for name, prop := range s.Properties {
		fullName := name
		if parent != "" {
			fullName = parent + "." + name
		}

		*params = append(*params, APIAutoFillParam{
			Name:     fullName,
			Position: "body",
			Type:     getParamType(prop, doc),
			Required: isRequired(s.Required, name),
		})

		parseBodyParams(fullName, prop, doc, visited, params)
	}
}

func isRequired(list []string, name string) bool {
	for _, v := range list {
		if v == name {
			return true
		}
	}
	return false
}

// uniqueParams deduplicates parameters (one entry per key)
func uniqueParams(params []APIAutoFillParam) []APIAutoFillParam {
	seen := make(map[string]bool)
	var result []APIAutoFillParam

	for _, p := range params {
		key := p.Position + ":" + p.Name
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}
	return result
}

// LogAPIStatisticsDetails logs per-API fields at Debug (--debug-log).
func LogAPIStatisticsDetails(list []APIStatisticsInfo) {
	log.Debug("swagger apis parsed", "count", len(list))

	for i, api := range list {
		log.Debug("interface",
			"index", i+1,
			"method", api.Method,
			"path", api.Path,
		)
		if len(api.Parameters) == 0 {
			log.Debug("parameters", "none", true)
			continue
		}

		for _, param := range api.Parameters {
			log.Debug("parameter",
				"name", param.Name,
				"in", param.Position,
				"type", param.Type,
				"required", param.Required,
			)
		}
	}
}
