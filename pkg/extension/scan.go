package extension

import (
	"fmt"

	"swagger-exp-knife4j/pkg/scanner"
)

// ScanPhase marks where a ScanHook runs in the single-scan pipeline (scanrun.RunSingle).
type ScanPhase string

const (
	// PhaseBeforeResolve runs before scanner.ResolveSwaggerJSONURL.
	PhaseBeforeResolve ScanPhase = "before_resolve"
	// PhaseAfterResolve runs after OpenAPI JSON URL is resolved.
	PhaseAfterResolve ScanPhase = "after_resolve"
	// PhaseAfterSaveAPIDocs runs after api-docs.json is written under output dir.
	PhaseAfterSaveAPIDocs ScanPhase = "after_save_api_docs"
	// PhaseAfterAnalyze runs after paths are parsed into APIStatisticsInfo list.
	PhaseAfterAnalyze ScanPhase = "after_analyze"
	// PhaseAfterProbe runs after AutoRequestAllAPI completes.
	PhaseAfterProbe ScanPhase = "after_probe"
	// PhaseBeforeWrite runs before CSV/JSONL/DB writers and custom ScanWriter hooks.
	PhaseBeforeWrite ScanPhase = "before_write"
	// PhaseAfterWrite runs after all writers succeed.
	PhaseAfterWrite ScanPhase = "after_write"
)

// ScanContext is passed to ScanHook and ScanWriter. Fields are filled progressively by scanrun.
type ScanContext struct {
	Phase ScanPhase

	InputURL  string
	OutputDir string
	DocsOnly  bool
	HTTP      *scanner.HTTPOptions

	ResolvedJSONURL string
	APIDocsPath     string
	Stats           []scanner.APIStatisticsInfo
	ProbeResults    []scanner.APIRequestResult
	HTTPMeta        scanner.HTTPRequestMeta

	// Abort, when non-nil, stops RunSingle and returns this error (hooks should not set unless intentional).
	Abort error
}

// ScanHook observes or mutates scan state at registered phases.
// Return a non-nil error to abort the scan; wrap with fmt.Errorf("my-hook: %w", err).
type ScanHook interface {
	// Name is a unique stable identifier (e.g. "acme.auth-enricher").
	Name() string
	// Phases lists phases this hook cares about; other phases are skipped.
	Phases() []ScanPhase
	// OnScan is invoked for each phase in Phases(). Read/write ScanContext fields allowed per phase table in module-development.md.
	OnScan(ctx *ScanContext) error
}

// ScanWriter persists or exports probe results (runs once per scan at PhaseBeforeWrite, before built-in writers).
type ScanWriter interface {
	Name() string
	Write(ctx *ScanContext) error
}

func phaseSet(phases []ScanPhase) map[ScanPhase]struct{} {
	m := make(map[ScanPhase]struct{}, len(phases))
	for _, p := range phases {
		m[p] = struct{}{}
	}
	return m
}

// RunScanHooks invokes all registered ScanHook instances that subscribe to phase.
// If ctx.Abort is set, returns it immediately without calling further hooks.
func RunScanHooks(phase ScanPhase, ctx *ScanContext) error {
	if ctx == nil {
		return fmt.Errorf("extension: ScanContext is nil")
	}
	ctx.Phase = phase
	if ctx.Abort != nil {
		return ctx.Abort
	}
	for _, h := range Default().ScanHooks() {
		if !hookWantsPhase(h, phase) {
			continue
		}
		if err := h.OnScan(ctx); err != nil {
			return fmt.Errorf("scan hook %q at %s: %w", h.Name(), phase, err)
		}
		if ctx.Abort != nil {
			return ctx.Abort
		}
	}
	return nil
}

// RunScanWriters invokes custom ScanWriter implementations before built-in CSV/JSONL/DB output.
func RunScanWriters(ctx *ScanContext) error {
	if ctx == nil {
		return fmt.Errorf("extension: ScanContext is nil")
	}
	for _, w := range Default().ScanWriters() {
		if err := w.Write(ctx); err != nil {
			return fmt.Errorf("scan writer %q: %w", w.Name(), err)
		}
	}
	return nil
}

func hookWantsPhase(h ScanHook, phase ScanPhase) bool {
	for _, p := range h.Phases() {
		if p == phase {
			return true
		}
	}
	return false
}
