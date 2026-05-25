// knife4j_proxy.go: proxies /api/* to the real target host during Knife4j debug (manifest TargetOrigin).
package reportserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"swagger-exp-knife4j/internal/islazy"
)

func knife4jTryBase(sessionID string) string {
	return "/knife4j/" + sessionID + "/try"
}

// TargetOrigin returns scheme://host for forwarding (from scan input_url when available).
func (s *Store) TargetOrigin(host string) string {
	for _, raw := range s.hostInputURLs()[host] {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" {
			continue
		}
		if u.Host != "" {
			return u.Scheme + "://" + u.Host
		}
		return u.Scheme + "://" + host
	}
	return "https://" + host
}

// TargetOriginForSession prefers manifest json_url/input_url for the session directory.
func (s *Store) TargetOriginForSession(sessionID string) string {
	if err := validateSessionID(sessionID); err == nil {
		base, err := filepath.Abs(s.cfg.APIDocPath)
		if err == nil {
			relDir := islazy.DecodeSessionDir(sessionID)
			manifestPath := filepath.Join(base, relDir, islazy.ManifestFileName)
			if data, err := os.ReadFile(manifestPath); err == nil {
				var manifest islazy.DocManifest
				if json.Unmarshal(data, &manifest) == nil {
					for _, raw := range []string{manifest.JSONURL, manifest.InputURL} {
						if origin := originFromURL(raw); origin != "" {
							return origin
						}
					}
				}
			}
		}
	}
	host, _, _ := islazy.ParseSessionMeta(sessionID)
	return s.TargetOrigin(host)
}

func originFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func (s *Server) handleKnife4jTryProxy(w http.ResponseWriter, r *http.Request) {
	// Do not forward OPTIONS to upstream; many APIs return 405 on preflight.
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sessionID := r.PathValue("sessionID")
	apiPath := r.PathValue("path")
	if sessionID == "" || strings.Contains(sessionID, "..") || strings.Contains(apiPath, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	host, _, _ := islazy.ParseSessionMeta(sessionID)
	if host == "" {
		http.Error(w, "unknown session host", http.StatusBadRequest)
		return
	}

	targetOrigin := s.store.TargetOriginForSession(sessionID)
	upstream, err := url.Parse(targetOrigin)
	if err != nil {
		http.Error(w, "invalid target", http.StatusInternalServerError)
		return
	}
	upstream.Path = "/" + strings.TrimPrefix(apiPath, "/")
	upstream.RawQuery = r.URL.RawQuery

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstream.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxyReq.ContentLength = r.ContentLength
	proxyReq.Host = upstream.Host

	for name, values := range r.Header {
		if hopByHopHeader(name) {
			continue
		}
		for _, v := range values {
			proxyReq.Header.Add(name, v)
		}
	}

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		if hopByHopHeader(name) {
			continue
		}
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func hopByHopHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailers", "transfer-encoding", "upgrade", "host":
		return true
	default:
		return false
	}
}
