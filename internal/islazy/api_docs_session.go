// api_docs_session.go: Derives host/scope/sessionID for the api-docs output directory, and handles manifest read/write operations.
package islazy

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const (
	// DefaultAPIScope is used when the swagger JSON sits at the site root.
	DefaultAPIScope = "default"
	sessionSep      = "__"
)

var swaggerDocSuffixes = []string{
	"/v3/api-docs",
	"/v2/api-docs",
	"/swagger.json",
	"/openapi.json",
	"/api-docs.json",
	"/api-docs",
}

// APIScopeDir derives a relative directory segment from the swagger JSON URL path.
// e.g. https://example.com/auth/v3/api-docs -> "auth"
func APIScopeDir(jsonURL string) string {
	u, err := url.Parse(jsonURL)
	if err != nil {
		return DefaultAPIScope
	}
	p := strings.TrimSuffix(u.Path, "/")
	for _, suffix := range swaggerDocSuffixes {
		if strings.HasSuffix(p, suffix) {
			p = strings.TrimSuffix(p, suffix)
			break
		}
	}
	p = strings.Trim(p, "/")
	if p == "" {
		return DefaultAPIScope
	}
	parts := strings.Split(p, "/")
	for i, part := range parts {
		part = SafePathSegment(part)
		if part == "" {
			part = "p"
		}
		parts[i] = part
	}
	return strings.Join(parts, "/")
}

// ScanSessionRelDir is the directory under output/, e.g. example.com/auth
func ScanSessionRelDir(host, scopeDir string) string {
	host = SafeFileName(host)
	if host == "" {
		host = "unknown"
	}
	if scopeDir == "" || scopeDir == DefaultAPIScope {
		return host
	}
	return filepath.Join(host, filepath.FromSlash(scopeDir))
}

// ScanAPIDocsPathForURL returns {baseDir}/{host}/{scope}/api-docs.json
func ScanAPIDocsPathForURL(baseDir, jsonURL string, scannedAt time.Time) string {
	host := hostFromURL(jsonURL)
	scope := APIScopeDir(jsonURL)
	return filepath.Join(baseDir, ScanSessionRelDir(host, scope), "api-docs.json")
}

// EncodeSessionID maps a relative output dir to a URL-safe session id (no slashes).
func EncodeSessionID(relDir string) string {
	relDir = filepath.ToSlash(relDir)
	return strings.ReplaceAll(relDir, "/", sessionSep)
}

// DecodeSessionDir maps session id back to a relative directory under output/.
func DecodeSessionDir(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	if legacy, ok := parseLegacySessionDir(sessionID); ok {
		return legacy
	}
	return filepath.FromSlash(strings.ReplaceAll(sessionID, sessionSep, "/"))
}

// ParseSessionMeta returns host, scope label, and scanned time from a session folder name.
func ParseSessionMeta(sessionID string) (host string, scope string, scannedAt time.Time) {
	dir := DecodeSessionDir(sessionID)
	if legacyHost, t, ok := parseLegacySessionHost(sessionID); ok {
		return legacyHost, "", t
	}
	dir = filepath.ToSlash(dir)
	parts := strings.Split(dir, "/")
	if len(parts) == 1 {
		return parts[0], "", time.Time{}
	}
	return parts[0], strings.Join(parts[1:], "/"), time.Time{}
}

func parseLegacySessionDir(sessionID string) (string, bool) {
	if _, _, ok := parseLegacySessionHost(sessionID); ok {
		return sessionID, true
	}
	return "", false
}

func parseLegacySessionHost(name string) (host string, scannedAt time.Time, ok bool) {
	parts := strings.SplitN(name, "_", 3)
	if len(parts) < 3 {
		return "", time.Time{}, false
	}
	t, err := time.ParseInLocation("20060102_150405", parts[0]+"_"+parts[1], time.Local)
	if err != nil {
		return "", time.Time{}, false
	}
	return parts[2], t, true
}

func hostFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}

// SafePathSegment allows path segment characters in scope names.
func SafePathSegment(s string) string {
	var builder strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '.', r == '-', r == '_':
			builder.WriteRune(r)
		}
	}
	if builder.Len() == 0 {
		return ""
	}
	return builder.String()
}

// DocManifest is stored beside api-docs.json for listing metadata.
type DocManifest struct {
	Host      string    `json:"host"`
	Scope     string    `json:"scope"`
	InputURL  string    `json:"input_url,omitempty"`
	JSONURL   string    `json:"json_url"`
	ScannedAt time.Time `json:"scanned_at"`
}

// ManifestFileName is the metadata file name in each session directory.
const ManifestFileName = "manifest.json"

// SessionLabel builds a human-readable label for UI.
func SessionLabel(host, scope string) string {
	if scope == "" || scope == DefaultAPIScope {
		return host
	}
	return fmt.Sprintf("%s / %s", host, scope)
}
