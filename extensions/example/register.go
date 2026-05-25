// Package example is a reference extension (not linked by default).
// Enable: add to main.go:  import _ "swagger-exp-knife4j/extensions/example"
package example

import "swagger-exp-knife4j/pkg/extension"

func init() {
	must(extension.RegisterScanHook(&urlLogHook{}))
	must(extension.RegisterCommand(&helloCommand{}))
}

func must(err error) {
	if err != nil {
		panic("extensions/example: " + err.Error())
	}
}
