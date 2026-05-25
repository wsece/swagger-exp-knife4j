// scan_jobs.go: background scan job queue and status for Web-submitted tasks.
package reportserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"swagger-exp-knife4j/pkg/log"
	"swagger-exp-knife4j/pkg/scanrun"
	"swagger-exp-knife4j/pkg/scanner"
)

const (
	scanTaskStatusPending   = "pending"
	scanTaskStatusRunning   = "running"
	scanTaskStatusCompleted = "completed"
	scanTaskStatusFailed    = "failed"

	maxScanURLsPerTask = 200
	maxConcurrentScans = 2
)

// ScanTask represents one scan job submitted from the Web UI.
type ScanTask struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"` // single | batch
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Total      int       `json:"total"`
	OK         int       `json:"ok"`
	Failed     int       `json:"failed"`
	URLs       []string  `json:"urls,omitempty"`
	Results    []scanrun.SingleResult `json:"results,omitempty"`
	Errors     []scanrun.FileTargetErr `json:"errors,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// ScanJobManager runs scans asynchronously inside the report server process.
type ScanJobManager struct {
	store    *Store
	httpOpts *scanner.HTTPOptions
	mu       sync.RWMutex
	tasks    map[string]*ScanTask
	sem      chan struct{}
}

// NewScanJobManager creates the job manager; results go to the store's DB and output directory.
func NewScanJobManager(store *Store, httpOpts *scanner.HTTPOptions) *ScanJobManager {
	if httpOpts == nil {
		httpOpts = &scanner.HTTPOptions{Parallel: 1}
	}
	return &ScanJobManager{
		store:    store,
		httpOpts: httpOpts,
		tasks:    make(map[string]*ScanTask),
		sem:      make(chan struct{}, maxConcurrentScans),
	}
}

// Enqueue validates URLs and starts a background scan; returns the task ID.
func (m *ScanJobManager) Enqueue(urls []string) (string, error) {
	urls = normalizeTargetURLs(urls)
	if len(urls) == 0 {
		return "", fmt.Errorf("at least one URL is required")
	}
	if len(urls) > maxScanURLsPerTask {
		return "", fmt.Errorf("too many URLs (max %d)", maxScanURLsPerTask)
	}

	taskType := "single"
	if len(urls) > 1 {
		taskType = "batch"
	}

	id := newScanTaskID()
	now := time.Now()
	task := &ScanTask{
		ID:        id,
		Type:      taskType,
		Status:    scanTaskStatusPending,
		CreatedAt: now,
		Total:     len(urls),
		URLs:      append([]string(nil), urls...),
	}

	m.mu.Lock()
	m.tasks[id] = task
	m.mu.Unlock()

	go m.run(task)

	return id, nil
}

func (m *ScanJobManager) run(task *ScanTask) {
	m.sem <- struct{}{}
	defer func() { <-m.sem }()

	m.mu.Lock()
	task.Status = scanTaskStatusRunning
	started := time.Now()
	task.StartedAt = &started
	m.mu.Unlock()

	cfg := m.store.Config()
	base := scanrun.SingleParams{
		OutputDir: cfg.APIDocPath,
		HTTP:      m.httpOpts,
		Writers: scanrun.WriterConfig{
			DbURI: cfg.DbURI,
		},
	}

	result, err := scanrun.RunURLs(task.URLs, base, nil)

	finished := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	task.FinishedAt = &finished
	if result != nil {
		task.OK = result.OK
		task.Failed = result.Failed
		task.Results = result.Results
		task.Errors = result.Errors
	}
	if err != nil && result != nil && result.OK > 0 {
		task.Status = scanTaskStatusCompleted
		task.Error = err.Error()
		log.Warn("scan task finished with errors", "id", task.ID, "err", err)
		return
	}
	if err != nil {
		task.Status = scanTaskStatusFailed
		if task.Error == "" {
			task.Error = err.Error()
		}
		log.Error("scan task failed", "id", task.ID, "err", err)
		return
	}
	task.Status = scanTaskStatusCompleted
	log.Info("scan task completed", "id", task.ID, "ok", task.OK, "failed", task.Failed)
}

// Get returns a read-only snapshot of a task.
func (m *ScanJobManager) Get(id string) (*ScanTask, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

// List returns all tasks newest-first by creation time.
func (m *ScanJobManager) List() []ScanTask {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ScanTask, 0, len(m.tasks))
	for _, t := range m.tasks {
		out = append(out, *t)
	}
	sortScanTasksByCreated(out)
	return out
}

func newScanTaskID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func normalizeTargetURLs(urls []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(urls))
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		if u == "" || strings.HasPrefix(u, "#") {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func sortScanTasksByCreated(tasks []ScanTask) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
}
