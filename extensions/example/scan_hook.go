package example

import (
	"swagger-exp-knife4j/pkg/extension"
	"swagger-exp-knife4j/pkg/log"
)

type urlLogHook struct{}

func (urlLogHook) Name() string { return "example.url-log" }

func (urlLogHook) Phases() []extension.ScanPhase {
	return []extension.ScanPhase{extension.PhaseAfterResolve}
}

func (urlLogHook) OnScan(ctx *extension.ScanContext) error {
	if ctx == nil {
		return nil
	}
	log.Info("example hook: resolved OpenAPI URL", "url", ctx.ResolvedJSONURL)
	return nil
}
