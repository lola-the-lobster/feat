// Package state manages per-feature workflow state in .feat/ directory.
// Each feature gets its own directory with state tracking files.
package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const StateDirName = ".feat"

// ValidStates are the allowed workflow states for a feature.
var ValidStates = []string{"scaffold", "fix", "build", "test", "done"}

// Manager handles state operations for a project.
type Manager struct {
	projectRoot string
	stateDir    string
}

// NewManager creates a state manager for the given project root.
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		stateDir:    filepath.Join(projectRoot, StateDirName),
	}
}

// Init creates the .feat/ directory structure if it doesn't exist.
func (m *Manager) Init() error {
	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	featuresDir := filepath.Join(m.stateDir, "features")
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		return fmt.Errorf("creating features directory: %w", err)
	}
	return nil
}

// Exists returns true if the .feat/ directory exists.
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.stateDir)
	return !os.IsNotExist(err)
}

// SanitizeFeaturePath converts "auth/login" to "auth-login" for filesystem use.
func SanitizeFeaturePath(featurePath string) string {
	return strings.ReplaceAll(featurePath, "/", "-")
}

// GetFeatureDir returns the path to a feature's state directory.
func (m *Manager) GetFeatureDir(featurePath string) string {
	sanitized := SanitizeFeaturePath(featurePath)
	return filepath.Join(m.stateDir, "features", sanitized)
}

// SetCurrent marks a feature as the active one.
func (m *Manager) SetCurrent(featurePath string) error {
	if err := m.Init(); err != nil {
		return err
	}
	currentPath := filepath.Join(m.stateDir, "current")
	return os.WriteFile(currentPath, []byte(featurePath+"\n"), 0644)
}

// GetCurrent returns the active feature path, or empty string if none.
func (m *Manager) GetCurrent() (string, error) {
	currentPath := filepath.Join(m.stateDir, "current")
	data, err := os.ReadFile(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading current feature: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// Clear removes the current feature state.
func (m *Manager) Clear() error {
	currentPath := filepath.Join(m.stateDir, "current")
	if err := os.Remove(currentPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing current file: %w", err)
	}
	return nil
}

// GetFeatureState returns the current workflow state for a feature.
// Returns "scaffold" if no state is set (default).
func (m *Manager) GetFeatureState(featurePath string) (string, error) {
	if err := m.Init(); err != nil {
		return "", err
	}

	featureDir := m.GetFeatureDir(featurePath)
	statePath := filepath.Join(featureDir, "state")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Default state for new features
			return "scaffold", nil
		}
		return "", fmt.Errorf("reading feature state: %w", err)
	}

	state := strings.TrimSpace(string(data))
	if state == "" {
		return "scaffold", nil
	}
	return state, nil
}

// SetFeatureState updates the workflow state for a feature.
func (m *Manager) SetFeatureState(featurePath string, state string) error {
	if err := m.Init(); err != nil {
		return err
	}

	// Validate state
	valid := false
	for _, s := range ValidStates {
		if s == state {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid state: %s", state)
	}

	featureDir := m.GetFeatureDir(featurePath)
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return fmt.Errorf("creating feature directory: %w", err)
	}

	statePath := filepath.Join(featureDir, "state")
	if err := os.WriteFile(statePath, []byte(state+"\n"), 0644); err != nil {
		return fmt.Errorf("writing feature state: %w", err)
	}

	return nil
}

// FormatState formats the current feature state for display.
func FormatState(featurePath string) string {
	if featurePath == "" {
		return "No active feature\n"
	}
	return fmt.Sprintf("Current feature: %s\n", featurePath)
}

// FormatFeatureStatus returns a formatted status string for a feature including its state.
func FormatFeatureStatus(featurePath string, state string) string {
	if featurePath == "" {
		return "No active feature\n"
	}
	return fmt.Sprintf("Feature: %s\nState: %s\n", featurePath, state)
}
