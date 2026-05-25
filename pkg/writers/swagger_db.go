// swagger_db.go: SwaggerDbWriter writes scan report rows to the DB with response SimHash grouping.
package writers

import (
	"sync"
	"time"

	"swagger-exp-knife4j/internal/islazy"
	"swagger-exp-knife4j/pkg/database"
	"swagger-exp-knife4j/pkg/models"
	"swagger-exp-knife4j/pkg/scanner"

	"gorm.io/gorm"
)

// SwaggerDbWriter appends Swagger scan results to the database and groups similar responses
type SwaggerDbWriter struct {
	URI     string
	conn    *gorm.DB
	mutex   sync.Mutex
	grouper *islazy.HashGrouper
}

// NewSwaggerDbWriter initializes the Swagger database writer
func NewSwaggerDbWriter(uri string, debug bool) (*SwaggerDbWriter, error) {
	conn, err := database.SwaggerConnection(uri, debug)
	if err != nil {
		return nil, err
	}

	w := &SwaggerDbWriter{
		URI:     uri,
		conn:    conn,
		grouper: islazy.NewHashGrouper(islazy.DefaultResponseHammingThreshold),
	}
	if err := w.grouper.LoadFromDB(conn, &models.SwaggerAPIRecord{}, "response_sim_hash", "response_group_id"); err != nil {
		return nil, err
	}
	return w, nil
}

// WriteRecords batch-writes API records (one -u scan yields many rows; safe to call repeatedly on same DB)
func (w *SwaggerDbWriter) WriteRecords(records []scanner.SwaggerReportRecord) error {
	if len(records) == 0 {
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	scannedAt := time.Now()

	return w.conn.Transaction(func(tx *gorm.DB) error {
		for _, rec := range records {
			packetJSON, _ := scanner.MarshalHTTPExchange(rec.Exchange)
			if packetJSON == "" {
				packetJSON, _ = scanner.MarshalHTTPExchange(scanner.BuildHTTPExchange(rec, scanner.APIRequestResult{
					Method: rec.Method, Path: rec.Path, Host: rec.Host,
					FullURL: rec.FullURL, FinalURL: rec.FullURL,
					RequestBody: rec.RequestBody, RequestParams: rec.RequestParams,
					StatusCode: rec.StatusCode, ContentType: rec.ContentType,
					Response: rec.Response, Error: rec.Error,
				}))
			}
			row := models.SwaggerAPIRecord{
				ScannedAt:      scannedAt,
				InputURL:       rec.SwaggerURL,
				Method:         rec.Method,
				Host:           rec.Host,
				API:            rec.Path,
				Cookie:         rec.Cookie,
				UserAgent:      rec.UserAgent,
				Authorization:  rec.Authorization,
				RequestBody:    rec.RequestBody,
				RequestParams:  rec.RequestParams,
				ResponseBody:   rec.Response,
				StatusCode:     rec.StatusCode,
				Failed:         rec.Failed,
				FailedReason:   rec.FailedReason,
				ParameterNames: rec.ParameterNames,
				PacketJSON:     packetJSON,
			}

			if !rec.Failed && rec.Response != "" {
				simHash := islazy.ResponseSimHash(rec.Response)
				groupID, hashStr, distance, err := w.grouper.Assign(tx, &models.SwaggerAPIRecord{}, "response_group_id", simHash)
				if err != nil {
					return err
				}
				row.ResponseGroupID = groupID
				row.ResponseSimHash = hashStr
				row.SimilarityDistance = distance
			}

			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
