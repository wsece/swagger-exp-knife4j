package extension

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

// Registry holds all registered extensions. Use Default() or package-level Register* helpers.
type Registry struct {
	mu sync.RWMutex

	scanHooks []ScanHook
	commands  []CommandExtension
	mcpTools  []MCPTool
	writers   []ScanWriter
}

var defaultRegistry = &Registry{}

// Default returns the process-wide extension registry.
func Default() *Registry {
	return defaultRegistry
}

// RegisterScanHook adds a scan pipeline hook (duplicate Name() returns error).
func RegisterScanHook(h ScanHook) error {
	return defaultRegistry.registerScanHook(h)
}

// RegisterCommand adds a CLI subcommand attached under the root command.
func RegisterCommand(c CommandExtension) error {
	return defaultRegistry.registerCommand(c)
}

// RegisterMCPTool adds an MCP tool when mcp serve starts.
func RegisterMCPTool(t MCPTool) error {
	return defaultRegistry.registerMCPTool(t)
}

// RegisterScanWriter adds a custom result writer invoked after probe, before built-in DB/CSV/JSONL.
func RegisterScanWriter(w ScanWriter) error {
	return defaultRegistry.registerScanWriter(w)
}

func (r *Registry) registerScanHook(h ScanHook) error {
	if h == nil {
		return fmt.Errorf("extension: ScanHook is nil")
	}
	name := h.Name()
	if name == "" {
		return fmt.Errorf("extension: ScanHook.Name() is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.scanHooks {
		if existing.Name() == name {
			return fmt.Errorf("extension: duplicate scan hook %q", name)
		}
	}
	r.scanHooks = append(r.scanHooks, h)
	return nil
}

func (r *Registry) registerCommand(c CommandExtension) error {
	if c == nil {
		return fmt.Errorf("extension: CommandExtension is nil")
	}
	cmd := c.CobraCommand()
	if cmd == nil {
		return fmt.Errorf("extension: CommandExtension.CobraCommand() returned nil")
	}
	if cmd.Use == "" {
		return fmt.Errorf("extension: command Use is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.commands {
		if existing.CobraCommand().Use == cmd.Use {
			return fmt.Errorf("extension: duplicate command %q", cmd.Use)
		}
	}
	r.commands = append(r.commands, c)
	return nil
}

func (r *Registry) registerMCPTool(t MCPTool) error {
	if t == nil {
		return fmt.Errorf("extension: MCPTool is nil")
	}
	name := t.Name()
	if name == "" {
		return fmt.Errorf("extension: MCPTool.Name() is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.mcpTools {
		if existing.Name() == name {
			return fmt.Errorf("extension: duplicate mcp tool %q", name)
		}
	}
	r.mcpTools = append(r.mcpTools, t)
	return nil
}

func (r *Registry) registerScanWriter(w ScanWriter) error {
	if w == nil {
		return fmt.Errorf("extension: ScanWriter is nil")
	}
	name := w.Name()
	if name == "" {
		return fmt.Errorf("extension: ScanWriter.Name() is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.writers {
		if existing.Name() == name {
			return fmt.Errorf("extension: duplicate scan writer %q", name)
		}
	}
	r.writers = append(r.writers, w)
	return nil
}

// ScanHooks returns registered hooks in registration order.
func (r *Registry) ScanHooks() []ScanHook {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ScanHook, len(r.scanHooks))
	copy(out, r.scanHooks)
	return out
}

// MCPTools returns registered MCP tools.
func (r *Registry) MCPTools() []MCPTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]MCPTool, len(r.mcpTools))
	copy(out, r.mcpTools)
	return out
}

// ScanWriters returns registered custom writers.
func (r *Registry) ScanWriters() []ScanWriter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ScanWriter, len(r.writers))
	copy(out, r.writers)
	return out
}

// AttachCommands adds all registered commands to root (typically swagger-exp-knife4j root).
func AttachCommands(root *cobra.Command) {
	if root == nil {
		return
	}
	for _, ext := range Default().commandsSnapshot() {
		root.AddCommand(ext.CobraCommand())
	}
}

func (r *Registry) commandsSnapshot() []CommandExtension {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]CommandExtension, len(r.commands))
	copy(out, r.commands)
	return out
}
