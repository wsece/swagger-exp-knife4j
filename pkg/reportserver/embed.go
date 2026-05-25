// embed.go: go:embed for static/ and Knife4j frontend assets.
package reportserver

import "embed"

//go:embed static/knife4j/doc.html static/knife4j/webjars/**
var knife4jFS embed.FS
