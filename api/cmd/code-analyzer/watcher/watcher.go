package watcher

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// IgnoredDirs contains directory names that should not be watched
var IgnoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".nix-profile": true,
	".nix-defexpr": true,
	".cache":       true,
	".vscode":      true,
	".idea":        true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".next":        true,
	"coverage":     true,
}

// IgnoredExtensions contains file extensions that should not trigger analysis
var IgnoredExtensions = map[string]bool{
	".log":   true,
	".tmp":   true,
	".swp":   true,
	".swo":   true,
	".pyc":   true,
	".pyo":   true,
	".o":     true,
	".a":     true,
	".so":    true,
	".dylib": true,
	".exe":   true,
	".dll":   true,
	".class": true,
	".jar":   true,
	".war":   true,
	".lock":  true,
}

// WorkspaceWatcher watches a workspace directory for file changes
type WorkspaceWatcher struct {
	workspacePath string
	workspaceName string
	watcher       *fsnotify.Watcher
	debouncer     *Debouncer
	onAnalyze     func(workspaceName string)
	logger        *zap.Logger
	mu            sync.Mutex
	running       bool
}

// NewWorkspaceWatcher creates a new workspace watcher
func NewWorkspaceWatcher(
	workspacePath string,
	workspaceName string,
	debounceDuration time.Duration,
	onAnalyze func(workspaceName string),
	logger *zap.Logger,
) *WorkspaceWatcher {
	w := &WorkspaceWatcher{
		workspacePath: workspacePath,
		workspaceName: workspaceName,
		onAnalyze:     onAnalyze,
		logger:        logger.With(zap.String("workspace", workspaceName)),
	}

	w.debouncer = NewDebouncer(debounceDuration, func() {
		w.logger.Info("Debounce timer fired, triggering analysis")
		w.onAnalyze(w.workspaceName)
	})

	return w
}

// Start begins watching the workspace directory
func (w *WorkspaceWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.mu.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	// Add all directories recursively
	if err := w.addDirsRecursively(w.workspacePath); err != nil {
		w.logger.Warn("Error adding directories", zap.Error(err))
	}

	w.logger.Info("Started watching workspace", zap.String("path", w.workspacePath))

	// Event loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				w.Stop()
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				w.handleEvent(event)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				w.logger.Error("Watcher error", zap.Error(err))
			}
		}
	}()

	return nil
}

// Stop stops watching the workspace
func (w *WorkspaceWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	w.debouncer.Cancel()

	if w.watcher != nil {
		w.watcher.Close()
		w.watcher = nil
	}

	w.logger.Info("Stopped watching workspace")
}

// IsRunning returns true if the watcher is running
func (w *WorkspaceWatcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// IsPending returns true if there's a pending analysis
func (w *WorkspaceWatcher) IsPending() bool {
	return w.debouncer.IsPending()
}

func (w *WorkspaceWatcher) addDirsRecursively(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		if d.IsDir() {
			name := d.Name()

			// Skip hidden directories (except root)
			if path != root && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}

			// Skip ignored directories
			if IgnoredDirs[name] {
				return filepath.SkipDir
			}

			// Add directory to watcher
			if err := w.watcher.Add(path); err != nil {
				w.logger.Debug("Failed to add directory", zap.String("path", path), zap.Error(err))
			}
		}

		return nil
	})
}

func (w *WorkspaceWatcher) handleEvent(event fsnotify.Event) {
	// Skip if file should be ignored
	if w.shouldIgnore(event.Name) {
		return
	}

	// Log the event
	w.logger.Debug("File event",
		zap.String("name", event.Name),
		zap.String("op", event.Op.String()),
	)

	// Handle directory creation - add to watcher
	if event.Op&fsnotify.Create == fsnotify.Create {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if !w.shouldIgnoreDir(filepath.Base(event.Name)) {
				w.watcher.Add(event.Name)
			}
		}
	}

	// Trigger debounced analysis for write/create/remove/rename events
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
		w.debouncer.Trigger()
	}
}

func (w *WorkspaceWatcher) shouldIgnore(path string) bool {
	name := filepath.Base(path)

	// Ignore hidden files
	if strings.HasPrefix(name, ".") {
		return true
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(name))
	if IgnoredExtensions[ext] {
		return true
	}

	// Check if in ignored directory
	for ignoredDir := range IgnoredDirs {
		if strings.Contains(path, string(filepath.Separator)+ignoredDir+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

func (w *WorkspaceWatcher) shouldIgnoreDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	return IgnoredDirs[name]
}

// Manager manages multiple workspace watchers
type Manager struct {
	workspacesPath   string
	debounceDuration time.Duration
	onAnalyze        func(workspaceName string)
	logger           *zap.Logger
	watchers         map[string]*WorkspaceWatcher
	mu               sync.RWMutex
}

// NewManager creates a new watcher manager
func NewManager(
	workspacesPath string,
	debounceDuration time.Duration,
	onAnalyze func(workspaceName string),
	logger *zap.Logger,
) *Manager {
	return &Manager{
		workspacesPath:   workspacesPath,
		debounceDuration: debounceDuration,
		onAnalyze:        onAnalyze,
		logger:           logger,
		watchers:         make(map[string]*WorkspaceWatcher),
	}
}

// StartWatching starts watching a workspace
func (m *Manager) StartWatching(ctx context.Context, workspaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.watchers[workspaceName]; exists {
		return nil // Already watching
	}

	workspacePath := filepath.Join(m.workspacesPath, workspaceName)

	// Check if workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return err
	}

	watcher := NewWorkspaceWatcher(
		workspacePath,
		workspaceName,
		m.debounceDuration,
		m.onAnalyze,
		m.logger,
	)

	if err := watcher.Start(ctx); err != nil {
		return err
	}

	m.watchers[workspaceName] = watcher
	m.logger.Info("Started watching workspace", zap.String("workspace", workspaceName))

	return nil
}

// StopWatching stops watching a workspace
func (m *Manager) StopWatching(workspaceName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if watcher, exists := m.watchers[workspaceName]; exists {
		watcher.Stop()
		delete(m.watchers, workspaceName)
		m.logger.Info("Stopped watching workspace", zap.String("workspace", workspaceName))
	}
}

// GetWatchedWorkspaces returns list of watched workspaces
func (m *Manager) GetWatchedWorkspaces() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workspaces := make([]string, 0, len(m.watchers))
	for name := range m.watchers {
		workspaces = append(workspaces, name)
	}
	return workspaces
}

// IsWatching returns true if workspace is being watched
func (m *Manager) IsWatching(workspaceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.watchers[workspaceName]
	return exists
}

// IsPending returns true if workspace has pending analysis
func (m *Manager) IsPending(workspaceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if watcher, exists := m.watchers[workspaceName]; exists {
		return watcher.IsPending()
	}
	return false
}

// StopAll stops all watchers
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, watcher := range m.watchers {
		watcher.Stop()
		delete(m.watchers, name)
	}
	m.logger.Info("Stopped all watchers")
}
