// openapi_normalize_test.go: tests for OpenAPI normalization helpers.
package reportserver

import (
	"encoding/json"
	"testing"
)

func TestNormalizeOpenAPISpec_addsServersAndSchemaTypes(t *testing.T) {
	raw := []byte(`{
		"openapi": "3.0.1",
		"components": {
			"schemas": {
				"ApiResult": {
					"type": "object",
					"properties": {
						"result": {
							"additionalProperties": false,
							"nullable": true
						}
					}
				}
			}
		}
	}`)
	out := NormalizeOpenAPISpec(raw, "example.com")
	var doc map[string]interface{}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["servers"]; !ok {
		t.Fatal("expected servers")
	}
	schemas := doc["components"].(map[string]interface{})["schemas"].(map[string]interface{})
	result := schemas["ApiResult"].(map[string]interface{})["properties"].(map[string]interface{})["result"].(map[string]interface{})
	if result["type"] != "object" {
		t.Fatalf("expected result.type=object, got %v", result["type"])
	}
}
