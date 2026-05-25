// scan_api.go: REST API for Web scan tasks (create, list, query status).
package reportserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxScanUploadBytes = 1 << 20 // 1 MiB

// handleScanTasksList GET /api/scan/tasks
func (s *Server) handleScanTasksList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("GET only"))
		return
	}
	writeJSON(w, s.jobs.List())
}

func (s *Server) handleScanTaskGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("GET only"))
		return
	}
	task, ok := s.jobs.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, fmt.Errorf("task not found"))
		return
	}
	writeJSON(w, task)
}

func (s *Server) handleScanTaskCreate(w http.ResponseWriter, r *http.Request) {
	urls, err := parseScanTargets(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	id, err := s.jobs.Enqueue(urls)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	task, _ := s.jobs.Get(id)
	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, task)
}

type scanCreateJSON struct {
	URL      string   `json:"url"`
	URLs     []string `json:"urls"`
	URLsText string   `json:"urls_text"`
}

func parseScanTargets(r *http.Request) ([]string, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		return parseScanMultipart(r)
	}
	return parseScanJSON(r)
}

func parseScanJSON(r *http.Request) ([]string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxScanUploadBytes))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty body")
	}

	var req scanCreateJSON
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	var urls []string
	if u := strings.TrimSpace(req.URL); u != "" {
		urls = append(urls, u)
	}
	urls = append(urls, req.URLs...)
	if t := strings.TrimSpace(req.URLsText); t != "" {
		urls = append(urls, splitTargetLines(t)...)
	}
	return urls, nil
}

func parseScanMultipart(r *http.Request) ([]string, error) {
	if err := r.ParseMultipartForm(maxScanUploadBytes); err != nil {
		return nil, fmt.Errorf("parse form: %w", err)
	}

	var urls []string
	if u := strings.TrimSpace(r.FormValue("url")); u != "" {
		urls = append(urls, u)
	}
	if t := strings.TrimSpace(r.FormValue("urls_text")); t != "" {
		urls = append(urls, splitTargetLines(t)...)
	}

	file, hdr, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		if hdr.Size > maxScanUploadBytes {
			return nil, fmt.Errorf("file too large (max %d bytes)", maxScanUploadBytes)
		}
		fromFile, readErr := readTargetLines(file)
		if readErr != nil {
			return nil, readErr
		}
		urls = append(urls, fromFile...)
	} else if err != http.ErrMissingFile {
		return nil, err
	}

	return urls, nil
}

func splitTargetLines(text string) []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	return out
}

func readTargetLines(r io.Reader) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
