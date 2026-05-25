// server.go: report Web server HTTP routes, CORS, and static/Knife4j assets.
package reportserver

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"swagger-exp-knife4j/internal/islazy"
	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/models"
)

//go:embed static/index.html
var indexHTML string

// ServerConfig is the report HTTP server listen address.
type ServerConfig struct {
	Host         string
	Port         int
	Knife4jProxy bool // when true (or ?proxy=1), route try-it-out via same-origin /knife4j/.../try
}

// Server wraps mux and Store, serving /api/* and Knife4j routes.
type Server struct {
	cfg   ServerConfig
	store *Store
	jobs  *ScanJobManager
	mux   *http.ServeMux
}

// NewServer registers all routes and returns *Server.
func NewServer(store *Store, cfg ServerConfig) *Server {
	s := &Server{
		cfg:   cfg,
		store: store,
		jobs:  NewScanJobManager(store, nil),
		mux:   http.NewServeMux(),
	}
	s.routes()
	return s
}

// ListenAndServe starts HTTP on cfg.Host:cfg.Port; blocks until error.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Print("report server listening", "url", "http://"+addr)
	return http.Serve(ln, s.cors(s.mux))
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /{$}", s.handleIndex)
	s.mux.HandleFunc("GET /api/hosts", s.handleHosts)
	s.mux.HandleFunc("GET /api/hosts/{host}/records", s.handleHostRecords)
	s.mux.HandleFunc("GET /api/sessions", s.handleSessions)
	s.mux.HandleFunc("GET /api/docs/{sessionID}/openapi.json", s.handleDocSpec)
	s.mux.HandleFunc("GET /api/docs/{sessionID}/preview", s.handleDocPreview)
	s.mux.HandleFunc("GET /api/records/{id}", s.handleRecord)
	s.mux.HandleFunc("GET /api/status", s.handleStatus)

	s.mux.HandleFunc("GET /api/scan/tasks", s.handleScanTasksList)
	s.mux.HandleFunc("POST /api/scan/tasks", s.handleScanTaskCreate)
	s.mux.HandleFunc("GET /api/scan/tasks/{id}", s.handleScanTaskGet)

	s.mux.HandleFunc("GET /knife4j-assets/{path...}", s.handleKnife4jStatic)
	// knife4j-doc.html loads lazy chunks from /webjars/... (site root)
	s.mux.HandleFunc("GET /webjars/{path...}", s.handleKnife4jStatic)
	// Legacy: lazy chunks under /knife4j/{sessionID}/webjars/...
	s.mux.HandleFunc("GET /knife4j/{sessionID}/webjars/{path...}", s.handleKnife4jSessionWebjars)
	s.mux.HandleFunc("GET /knife4j-doc.html", s.handleKnife4jDocFlat)
	s.mux.HandleFunc("GET /knife4j-swagger-config", s.handleKnife4jSwaggerConfigFlat)
	// Fallback when Knife4j requests /v3/api-docs/* from site root (flat doc page)
	s.mux.HandleFunc("GET /v3/api-docs/swagger-config", s.handleKnife4jSwaggerConfigQuery)
	s.mux.HandleFunc("GET /v3/api-docs", s.handleKnife4jOpenAPIQuery)
	s.mux.HandleFunc("GET /knife4j/{sessionID}/doc.html", s.handleKnife4jDoc)
	s.mux.HandleFunc("GET /knife4j/{sessionID}/v3/api-docs/swagger-config", s.handleKnife4jSwaggerConfig)
	s.mux.HandleFunc("GET /knife4j/{sessionID}/v3/api-docs", s.handleKnife4jOpenAPI)
	// Flat spec URL so Knife4j basePath stays "/" (avoids /knife4j/{session} prefix in UI)
	s.mux.HandleFunc("GET /knife4j-openapi.json", s.handleKnife4jOpenAPIFlat)
	// Knife4j debug proxy (same-origin; avoids cross-origin OPTIONS preflight)
	s.mux.HandleFunc("/knife4j/{sessionID}/try/{path...}", s.handleKnife4jTryProxy)
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With, X-CSRF-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	hosts, _ := s.store.ListHosts()
	sessions, _ := s.store.ListDocSessions()
	writeJSON(w, map[string]interface{}{
		"db_uri":       s.store.cfg.DbURI,
		"api_doc_path": s.store.cfg.APIDocPath,
		"host_count":   len(hosts),
		"session_count": len(sessions),
	})
}

func (s *Server) handleHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := s.store.ListHosts()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, hosts)
}

func (s *Server) handleHostRecords(w http.ResponseWriter, r *http.Request) {
	host := r.PathValue("host")
	if host == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("host required"))
		return
	}
	records, err := s.store.ListRecordsByHost(host)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	clusters := similarityClustersForRecords(records)
	out := make([]RecordJSON, len(records))
	for i, rec := range records {
		c := 0
		if i < len(clusters) {
			c = clusters[i]
		}
		out[i] = ToRecordJSON(rec, c)
	}
	writeJSON(w, out)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.store.ListDocSessions()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, sessions)
}

func (s *Server) handleDocSpec(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	data, err := s.store.ReadDocSpec(sessionID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}
	host, _, _ := islazy.ParseSessionMeta(sessionID)
	data = NormalizeOpenAPISpec(data, host)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(PrettyJSON(data))
}

func (s *Server) handleRecord(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	rec, err := s.store.GetRecord(uint(id))
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}
	clusters := similarityClustersForRecords([]models.SwaggerAPIRecord{*rec})
	c := 0
	if len(clusters) > 0 {
		c = clusters[0]
	}
	writeJSON(w, ToRecordJSON(*rec, c))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(jsonEmptySlice(v))
}

// jsonEmptySlice avoids nil slices encoding as JSON null (breaks frontend .length).
func jsonEmptySlice(v interface{}) interface{} {
	switch x := v.(type) {
	case []HostSummary:
		if x == nil {
			return []HostSummary{}
		}
	case []DocSession:
		if x == nil {
			return []DocSession{}
		}
	case []RecordJSON:
		if x == nil {
			return []RecordJSON{}
		}
	case []ScanTask:
		if x == nil {
			return []ScanTask{}
		}
	}
	return v
}

func writeJSONError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
