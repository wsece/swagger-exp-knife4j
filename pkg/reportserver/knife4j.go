// knife4j.go: embedded Knife4j static assets and doc.html entry handling.
package reportserver

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
)

// DocPreview fuzzy preview payload for api-docs
type DocPreview struct {
	SessionID  string   `json:"session_id"`
	Title      string   `json:"title"`
	Version    string   `json:"version"`
	PathCount  int      `json:"path_count"`
	Paths      []string `json:"paths"`
	Excerpt    string   `json:"excerpt"`
	SpecURL    string   `json:"spec_url"`
	RenderURL  string   `json:"render_url"`
}

// BuildDocPreview builds preview data from api-docs.json
func (s *Store) BuildDocPreview(sessionID string) (*DocPreview, error) {
	data, err := s.ReadDocSpec(sessionID)
	if err != nil {
		return nil, err
	}

	var doc struct {
		Info struct {
			Title   string `json:"title"`
			Version string `json:"version"`
		} `json:"info"`
		Paths map[string]interface{} `json:"paths"`
	}
	_ = json.Unmarshal(data, &doc)

	paths := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		paths = append(paths, p)
	}
	if len(paths) > 24 {
		paths = paths[:24]
	}

	excerpt := string(data)
	if len(excerpt) > 2048 {
		excerpt = excerpt[:2048] + "..."
	}

	return &DocPreview{
		SessionID: sessionID,
		Title:     doc.Info.Title,
		Version:   doc.Info.Version,
		PathCount: len(doc.Paths),
		Paths:     paths,
		Excerpt:   excerpt,
		SpecURL:   "/api/docs/" + sessionID + "/openapi.json",
		RenderURL: knife4jDocURL(sessionID),
	}, nil
}

func knife4jDocURL(sessionID string) string {
	return "/knife4j-doc.html?session=" + url.QueryEscape(sessionID)
}

func knife4jSpecURL(sessionID string) string {
	return "/knife4j-openapi.json?session=" + url.QueryEscape(sessionID)
}

func knife4jSwaggerConfigURL(sessionID string) string {
	return "/knife4j-swagger-config?session=" + url.QueryEscape(sessionID)
}

func knife4jSwaggerConfig(sessionID, targetOrigin string) map[string]interface{} {
	specPath := knife4jSpecURL(sessionID)
	cfg := map[string]interface{}{
		"urls": []map[string]string{
			{
				"name":           sessionID,
				"url":            specPath,
				"swaggerVersion": "3.0",
				"location":       specPath,
			},
		},
		"validatorUrl":   "",
		"basePath":       "/",
		"rootPath":       "/",
	}
	if targetOrigin != "" {
		cfg["host"] = strings.TrimPrefix(strings.TrimPrefix(targetOrigin, "https://"), "http://")
		if strings.HasPrefix(targetOrigin, "http://") {
			cfg["schemes"] = []string{"http", "https"}
		} else {
			cfg["schemes"] = []string{"https", "http"}
		}
	}
	return cfg
}

func (s *Server) handleKnife4jDoc(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	if sessionID == "" || strings.Contains(sessionID, "..") {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, knife4jDocURL(sessionID), http.StatusFound)
}

func (s *Server) handleKnife4jDocFlat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" || strings.Contains(sessionID, "..") || strings.ContainsAny(sessionID, `/\`) {
		http.Error(w, "session query required", http.StatusBadRequest)
		return
	}
	s.writeKnife4jDocHTML(w, sessionID)
}

func (s *Server) handleKnife4jSwaggerConfigQuery(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = sessionFromKnife4jReferer(r)
	}
	if sessionID == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("session required"))
		return
	}
	s.writeKnife4jSwaggerConfig(w, sessionID)
}

func (s *Server) handleKnife4jOpenAPIQuery(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = sessionFromKnife4jReferer(r)
	}
	if sessionID == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("session required"))
		return
	}
	s.writeKnife4jOpenAPI(w, sessionID)
}

func sessionFromKnife4jReferer(r *http.Request) string {
	ref := r.Referer()
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	if q := u.Query().Get("session"); q != "" {
		return q
	}
	if strings.HasSuffix(u.Path, "/knife4j-doc.html") || u.Path == "/knife4j-doc.html" {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "knife4j" && parts[len(parts)-1] == "doc.html" {
		return parts[1]
	}
	return ""
}

func (s *Server) writeKnife4jDocHTML(w http.ResponseWriter, sessionID string) {
	data, err := knife4jFS.ReadFile("static/knife4j/doc.html")
	if err != nil {
		http.Error(w, "doc template missing", http.StatusInternalServerError)
		return
	}
	targetOrigin := s.store.TargetOriginForSession(sessionID)
	html := string(data)
	html = strings.ReplaceAll(html, "__SESSION_ID__", sessionID)
	html = strings.ReplaceAll(html, "__TARGET_ORIGIN__", targetOrigin)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleDocPreview(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	preview, err := s.store.BuildDocPreview(sessionID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, preview)
}

func (s *Server) handleKnife4jSwaggerConfig(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	if sessionID == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("session required"))
		return
	}
	s.writeKnife4jSwaggerConfig(w, sessionID)
}

func (s *Server) handleKnife4jSwaggerConfigFlat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" || strings.Contains(sessionID, "..") || strings.ContainsAny(sessionID, `/\`) {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("session query required"))
		return
	}
	s.writeKnife4jSwaggerConfig(w, sessionID)
}

func (s *Server) writeKnife4jSwaggerConfig(w http.ResponseWriter, sessionID string) {
	targetOrigin := s.store.TargetOriginForSession(sessionID)
	writeJSON(w, knife4jSwaggerConfig(sessionID, targetOrigin))
}

func (s *Server) handleKnife4jOpenAPIFlat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" || strings.Contains(sessionID, "..") || strings.ContainsAny(sessionID, `/\`) {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("session query required"))
		return
	}
	s.writeKnife4jOpenAPI(w, sessionID)
}

func (s *Server) handleKnife4jOpenAPI(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	if sessionID == "" {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("session required"))
		return
	}
	s.writeKnife4jOpenAPI(w, sessionID)
}

func (s *Server) writeKnife4jOpenAPI(w http.ResponseWriter, sessionID string) {
	data, err := s.store.ReadDocSpec(sessionID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}
	// Route Knife4j "try" requests through same-origin proxy to avoid browser CORS preflight (OPTIONS).
	data = NormalizeOpenAPISpec(data, Knife4jProxyServerURL(sessionID))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (s *Server) handleKnife4jStatic(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	serveKnife4jWebjar(w, r, path)
}

func (s *Server) handleKnife4jSessionWebjars(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	if sessionID == "" || strings.Contains(sessionID, "..") {
		http.NotFound(w, r)
		return
	}
	path := r.PathValue("path")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	serveKnife4jWebjar(w, r, path)
}

func serveKnife4jWebjar(w http.ResponseWriter, r *http.Request, path string) {
	if strings.Contains(path, "..") {
		http.NotFound(w, r)
		return
	}
	f, err := knife4jFS.Open("static/knife4j/webjars/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(knife4jFS, "static/knife4j/webjars/"+path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", mimeByExt(path))
	_, _ = w.Write(data)
}

func mimeByExt(path string) string {
	switch {
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
