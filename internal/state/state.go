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

// Manager handles state operations for a project.
type Manager struct {
	projectRoot string
	stateDir    string
	workflow    []string // Workflow steps from manifest
}

// NewManager creates a state manager for the given project root.
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		stateDir:    filepath.Join(projectRoot, StateDirName),
		workflow:    nil,
	}
}

// SetWorkflow sets the workflow steps for the manager.
// Used to provide default step when no step is set.
func (m *Manager) SetWorkflow(workflow []string) {
	m.workflow = workflow
}

// getDefaultStep returns the first step in the workflow, or empty string if no workflow set.
func (m *Manager) getDefaultStep() string {
	if len(m.workflow) > 0 {
		return m.workflow[0]
	}
	return ""
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

// GetFeatureStep returns the current workflow step for a feature.
// Returns the first step in the workflow if no step is set (default).
func (m *Manager) GetFeatureStep(featurePath string) (string, error) {
	if err := m.Init(); err != nil {
		return "", err
	}

	featureDir := m.GetFeatureDir(featurePath)
	stepPath := filepath.Join(featureDir, "step")

	data, err := os.ReadFile(stepPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Default step for new features
			return m.getDefaultStep(), nil
		}
		return "", fmt.Errorf("reading feature step: %w", err)
	}

	step := strings.TrimSpace(string(data))
	if step == "" {
		return m.getDefaultStep(), nil
	}
	return step, nil
}

// SetFeatureStep updates the workflow step for a feature.
func (m *Manager) SetFeatureStep(featurePath string, step string) error {
	if err := m.Init(); err != nil {
		return err
	}

	featureDir := m.GetFeatureDir(featurePath)
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return fmt.Errorf("creating feature directory: %w", err)
	}

	stepPath := filepath.Join(featureDir, "step")
	if err := os.WriteFile(stepPath, []byte(step+"\n"), 0644); err != nil {
		return fmt.Errorf("writing feature step: %w", err)
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

// FormatFeatureStatus returns a formatted status string for a feature including its step.
func FormatFeatureStatus(featurePath string, step string) string {
	if featurePath == "" {
		return "No active feature\n"
	}
	return fmt.Sprintf("Feature: %s\nStep: %s\n", featurePath, step)
}
