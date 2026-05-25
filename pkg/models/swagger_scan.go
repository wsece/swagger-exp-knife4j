// swagger_scan.go: GORM model SwaggerAPIRecord (table swagger_api_records).
package models

import "time"

// SwaggerAPIRecord is one API scan/request row (many URLs can share one database)
type SwaggerAPIRecord struct {
	ID uint `json:"id" gorm:"primarykey"`

	ScannedAt time.Time `json:"when" gorm:"index"`
	InputURL  string    `json:"input_url" gorm:"index"`

	Method string `json:"method" gorm:"index"`
	Host   string `json:"host" gorm:"index"`
	API    string `json:"api" gorm:"index"`

	Cookie        string `json:"cookie"`
	UserAgent     string `json:"user_agent"`
	Authorization string `json:"authorization"`

	RequestBody   string `json:"request_body"`
	RequestParams string `json:"request_params"`
	ResponseBody  string `json:"response_body"`
	StatusCode    int    `json:"status_code" gorm:"index"`

	Failed       bool   `json:"failed" gorm:"index"`
	FailedReason string `json:"failed_reason"`

	SimilarityDistance int    `json:"similarity_distance" gorm:"index"`
	ResponseGroupID    uint   `json:"response_group_id" gorm:"index"`
	ResponseSimHash    string `json:"response_sim_hash" gorm:"index"`

	ParameterNames string `json:"parameter_names"`

	// PacketJSON Burp-style request/response capture (meta + request + response)
	PacketJSON string `json:"packet_json,omitempty" gorm:"type:text"`
}
