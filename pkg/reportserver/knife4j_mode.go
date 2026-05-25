// knife4j_mode.go: direct vs same-origin proxy mode for Knife4j debug requests.
package reportserver

import (
	"net/http"
	"strings"
)

// knife4jUseProxy reports whether try-it-out requests go through /knife4j/{session}/try (CORS workaround).
func (s *Server) knife4jUseProxy(r *http.Request) bool {
	if r != nil {
		switch strings.ToLower(r.URL.Query().Get("proxy")) {
		case "1", "true", "yes":
			return true
		case "0", "false", "no":
			return false
		}
	}
	return s.cfg.Knife4jProxy
}
