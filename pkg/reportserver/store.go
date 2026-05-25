// store.go: aggregates DB records and on-disk api-docs sessions; powers /api/* queries and record detail.
package reportserver

import (
	"encoding/json"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"swagger-exp-knife4j/internal/islazy"
	"swagger-exp-knife4j/pkg/database"
	"swagger-exp-knife4j/pkg/models"
	"swagger-exp-knife4j/pkg/scanner"

	"gorm.io/gorm"
)

// Config for the report web server
type Config struct {
	DbURI      string
	APIDocPath string
}

// Store aggregates DB records and filesystem api-docs sessions
type Store struct {
	cfg  Config
	conn *gorm.DB
}

// HostSummary scan stats grouped by host
type HostSummary struct {
	Host         string    `json:"host"`
	RecordCount  int       `json:"record_count"`
	LastScanned  time.Time `json:"last_scanned"`
	InputURLs    []string  `json:"input_urls"`
	RiskCount    int       `json:"risk_count"`
	DocSessionID string    `json:"doc_session_id,omitempty"`
}

// DocSession api-docs bundle under output/{host}/{scope}/api-docs.json
type DocSession struct {
	ID          string    `json:"id"`
	Host        string    `json:"host"`
	Scope       string    `json:"scope,omitempty"`
	Label       string    `json:"label"`
	ScannedAt   time.Time `json:"scanned_at"`
	InputURL    string    `json:"input_url,omitempty"`
	JSONURL     string    `json:"json_url,omitempty"`
	RecordCount int       `json:"record_count"`
	SpecURL     string    `json:"spec_url"`
}

// NewStore opens DB and prepares store
func NewStore(cfg Config) (*Store, error) {
	if abs, err := filepath.Abs(cfg.APIDocPath); err == nil {
		cfg.APIDocPath = abs
	}
	conn, err := database.SwaggerConnection(cfg.DbURI, false)
	if err != nil {
		return nil, err
	}
	return &Store{cfg: cfg, conn: conn}, nil
}

// Config returns store configuration
func (s *Store) Config() Config {
	return s.cfg
}

// ListHosts returns hosts with summary from DB, linked to doc sessions when possible
func (s *Store) ListHosts() ([]HostSummary, error) {
	var records []models.SwaggerAPIRecord
	if err := s.conn.Order("scanned_at desc").Find(&records).Error; err != nil {
		return nil, err
	}

	hostRecords := s.hostRecordCounts()

	sessions, _ := s.ListDocSessions()
	sessionByHost := make(map[string]DocSession)
	for _, sess := range sessions {
		existing, ok := sessionByHost[sess.Host]
		if !ok || sess.ScannedAt.After(existing.ScannedAt) {
			sessionByHost[sess.Host] = sess
		}
	}

	byHost := make(map[string]*HostSummary)
	urlSeen := make(map[string]map[string]struct{})

	for _, r := range records {
		host := r.Host
		if host == "" {
			host = "unknown"
		}
		sum, ok := byHost[host]
		if !ok {
			sum = &HostSummary{Host: host}
			byHost[host] = sum
			urlSeen[host] = make(map[string]struct{})
		}
		sum.RecordCount++
		if r.ScannedAt.After(sum.LastScanned) {
			sum.LastScanned = r.ScannedAt
		}
		if !r.Failed && r.StatusCode != 0 && r.StatusCode != 401 {
			sum.RiskCount++
		}
		if r.InputURL != "" {
			urlSeen[host][r.InputURL] = struct{}{}
		}
	}

	out := make([]HostSummary, 0, len(byHost))
	for host, sum := range byHost {
		for u := range urlSeen[host] {
			sum.InputURLs = append(sum.InputURLs, u)
		}
		sort.Strings(sum.InputURLs)
		if sess, ok := sessionByHost[host]; ok {
			sum.DocSessionID = sess.ID
		}
		out = append(out, *sum)
	}

	for _, sess := range sessions {
		if _, ok := byHost[sess.Host]; ok {
			continue
		}
		out = append(out, HostSummary{
			Host:         sess.Host,
			RecordCount:  hostRecords[sess.Host],
			LastScanned:  sess.ScannedAt,
			DocSessionID: sess.ID,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].LastScanned.After(out[j].LastScanned)
	})
	return out, nil
}

// ListRecordsByHost returns API records for a host
func (s *Store) ListRecordsByHost(host string) ([]models.SwaggerAPIRecord, error) {
	var records []models.SwaggerAPIRecord
	q := s.conn.Where("host = ?", host).Order("scanned_at desc, id desc")
	if err := q.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// ListDocSessions discovers all api-docs.json under output/ (one entry per host+scope).
func (s *Store) ListDocSessions() ([]DocSession, error) {
	base, err := filepath.Abs(s.cfg.APIDocPath)
	if err != nil {
		return nil, err
	}

	hostRecords := s.hostRecordCounts()

	sessions := make([]DocSession, 0)
	err = filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "api-docs.json" {
			return nil
		}
		dir := filepath.Dir(path)
		rel, err := filepath.Rel(base, dir)
		if err != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		sessionID := islazy.EncodeSessionID(rel)

		host, scope, _ := islazy.ParseSessionMeta(sessionID)
		var manifest islazy.DocManifest
		if data, err := os.ReadFile(filepath.Join(dir, islazy.ManifestFileName)); err == nil {
			_ = json.Unmarshal(data, &manifest)
		}
		if manifest.Host != "" {
			host = manifest.Host
		}
		if manifest.Scope != "" {
			scope = manifest.Scope
		}
		scannedAt := manifest.ScannedAt
		if scannedAt.IsZero() {
			if info, err := d.Info(); err == nil {
				scannedAt = info.ModTime()
			}
		}

		sess := DocSession{
			ID:          sessionID,
			Host:        host,
			Scope:       scope,
			Label:       islazy.SessionLabel(host, scope),
			ScannedAt:   scannedAt,
			InputURL:    manifest.InputURL,
			JSONURL:     manifest.JSONURL,
			RecordCount: hostRecords[host],
			SpecURL:     "/api/docs/" + url.PathEscape(sessionID) + "/openapi.json",
		}
		sessions = append(sessions, sess)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return []DocSession{}, nil
		}
		return nil, err
	}

	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].Host != sessions[j].Host {
			return sessions[i].Host < sessions[j].Host
		}
		if sessions[i].Scope != sessions[j].Scope {
			return sessions[i].Scope < sessions[j].Scope
		}
		return sessions[i].ScannedAt.After(sessions[j].ScannedAt)
	})
	return sessions, nil
}

// ReadDocSpec loads api-docs.json bytes for a session id
func (s *Store) ReadDocSpec(sessionID string) ([]byte, error) {
	if err := validateSessionID(sessionID); err != nil {
		return nil, err
	}
	base, err := filepath.Abs(s.cfg.APIDocPath)
	if err != nil {
		return nil, err
	}
	relDir := islazy.DecodeSessionDir(sessionID)
	path := filepath.Join(base, relDir, "api-docs.json")
	return os.ReadFile(path)
}

func validateSessionID(sessionID string) error {
	if sessionID == "" || strings.Contains(sessionID, "..") {
		return os.ErrNotExist
	}
	return nil
}

// GetRecord returns one record by id
func (s *Store) GetRecord(id uint) (*models.SwaggerAPIRecord, error) {
	var rec models.SwaggerAPIRecord
	if err := s.conn.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Store) hostRecordCounts() map[string]int {
	var rows []struct {
		Host  string
		Count int
	}
	_ = s.conn.Model(&models.SwaggerAPIRecord{}).
		Select("host, count(*) as count").
		Group("host").
		Scan(&rows).Error
	m := make(map[string]int)
	for _, r := range rows {
		m[r.Host] = r.Count
	}
	return m
}

func (s *Store) hostInputURLs() map[string][]string {
	var records []models.SwaggerAPIRecord
	_ = s.conn.Select("host", "input_url").Find(&records).Error
	m := make(map[string]map[string]struct{})
	for _, r := range records {
		if r.Host == "" || r.InputURL == "" {
			continue
		}
		if m[r.Host] == nil {
			m[r.Host] = make(map[string]struct{})
		}
		m[r.Host][r.InputURL] = struct{}{}
	}
	out := make(map[string][]string)
	for host, set := range m {
		for u := range set {
			out[host] = append(out[host], u)
		}
		sort.Strings(out[host])
	}
	return out
}

// RecordJSON for API responses with computed fields
type RecordJSON struct {
	models.SwaggerAPIRecord
	UnauthorizedRisk    bool                 `json:"unauthorized_risk"`
	Packet              scanner.HTTPExchange `json:"packet"`
	SimilarityCluster   int                  `json:"similarity_cluster"`
}

func similarityClustersForRecords(records []models.SwaggerAPIRecord) []int {
	n := len(records)
	if n == 0 {
		return nil
	}
	groupIDs := make([]uint, n)
	distances := make([]int, n)
	hashes := make([]string, n)
	bodies := make([]string, n)
	ties := make([]string, n)
	for i, r := range records {
		groupIDs[i] = r.ResponseGroupID
		distances[i] = r.SimilarityDistance
		hashes[i] = r.ResponseSimHash
		bodies[i] = r.ResponseBody
		ties[i] = r.Method + " " + r.API
	}
	return islazy.AssignResponseSimilarityClusters(groupIDs, distances, hashes, bodies, ties, 0)
}

func ToRecordJSON(r models.SwaggerAPIRecord, similarityCluster int) RecordJSON {
	risk := !r.Failed && r.StatusCode != 0 && r.StatusCode != 401
	packet, err := scanner.UnmarshalHTTPExchange(r.PacketJSON)
	if err != nil || r.PacketJSON == "" {
		packet = scanner.HTTPExchangeFromLegacyRecord(
			r.InputURL, r.Method, r.Host, r.API,
			r.Cookie, r.UserAgent, r.Authorization,
			r.RequestBody, r.RequestParams, r.ResponseBody,
			r.StatusCode, r.Failed, r.FailedReason,
		)
		packet.Meta.UnauthorizedRisk = risk
	}
	if r.ResponseSimHash == "" && r.ResponseBody != "" {
		if h := islazy.ResponseSimHash(r.ResponseBody); len(h) > 0 {
			r.ResponseSimHash = islazy.FormatResponseSimHash(h)
		}
	}
	return RecordJSON{
		SwaggerAPIRecord:  r,
		UnauthorizedRisk:  risk,
		Packet:            packet,
		SimilarityCluster: similarityCluster,
	}
}

func PrettyJSON(data []byte) []byte {
	var v interface{}
	if json.Unmarshal(data, &v) != nil {
		return data
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return data
	}
	return out
}
