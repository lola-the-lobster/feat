package main

import (
	exitcodes "github.com/lola-the-lobster/feat/internal/errors"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lola-the-lobster/feat/internal/formatter"
	"github.com/lola-the-lobster/feat/internal/loader"
	"github.com/lola-the-lobster/feat/internal/manifest"
	"github.com/lola-the-lobster/feat/internal/split"
	"github.com/lola-the-lobster/feat/internal/state"
	"github.com/lola-the-lobster/feat/internal/tree"
)

var (
	version = "dev"
	commit  = "unknown"
)

// Global flags
var (
	jsonOutput bool
)

func main() {
	// Parse global flags before command
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			// Remove --json from args
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(exitcodes.ExitGeneralError)
	}

	// Handle version flag
	if args[0] == "-v" || args[0] == "--version" || args[0] == "version" {
		if jsonOutput {
			fmt.Printf(`{"version": "%s", "commit": "%s"}`+"\n", version, commit)
		} else {
			fmt.Printf("feat version %s (commit: %s)\n", version, commit)
		}
		os.Exit(exitcodes.ExitSuccess)
	}

	// Handle help flag
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printUsage()
		os.Exit(exitcodes.ExitSuccess)
	}

	command := args[0]
	os.Args = append([]string{os.Args[0]}, args...)

	switch command {
	case "init":
		if err := runInit(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	case "list":
		if err := runList(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	case "parse":
		if err := runParse(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	case "split":
		if err := runSplit(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	case "status":
		if err := runStatus(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	case "validate":
		if err := runValidate(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	case "work":
		if err := runWork(); err != nil {
			printError(err, exitcodes.ExitGeneralError)
		}
	default:
		err := fmt.Errorf("unknown command: %s", command)
		printError(err, exitcodes.ExitGeneralError)
	}
}

// printError prints an error in text or JSON format depending on --json flag.
func printError(err error, code int) {
	if jsonOutput {
		output := map[string]interface{}{
			"error": err.Error(),
			"code":  code,
		}
		data, _ := json.Marshal(output)
		fmt.Fprintln(os.Stderr, string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(code)
}

func printUsage() {
	fmt.Println("feat - Feature-centric context management for agentic coding")
	fmt.Println()
	fmt.Println("Usage: feat <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init              Create a new feat.yaml manifest")
	fmt.Println("  list              Show feature tree")
	fmt.Println("  parse             Parse feat.yaml and dump structure")
	fmt.Println("  split <parent> <name>  Create a new feature")
	fmt.Println("  status            Show current feature context")
	fmt.Println("  validate          Check manifest for issues")
	fmt.Println("  work <feature>    Load a feature's context")
	fmt.Println("  version           Show version information")
	fmt.Println("  help              Show this help message")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  --json            Output in JSON format")
	fmt.Println("  -f <path>         Path to manifest file (default: feat.yaml)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  feat init                    # Create new manifest")
	fmt.Println("  feat list                    # Show all features")
	fmt.Println("  feat list --json             # Show features as JSON")
	fmt.Println("  feat work auth/login         # Work on auth/login feature")
	fmt.Println("  feat split auth login-v2     # Create auth/login-v2 feature")
	fmt.Println("  feat status                  # Show current feature")
	fmt.Println("  feat validate                # Check for issues")
}

func runInit() error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	var manifestPath string
	var projectName string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	fs.StringVar(&projectName, "name", "my-project", "Project name")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Check if already exists
	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("manifest already exists: %s", absPath)
	}

	if err := manifest.Init(absPath, projectName); err != nil {
		return fmt.Errorf("creating manifest: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"manifest": absPath,
			"project":  projectName,
			"message":  "Created manifest",
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Created manifest: %s\n", absPath)
		fmt.Printf("Project name: %s\n", projectName)
		fmt.Println("Add features to get started:")
		fmt.Println("  feat split \"\" my-feature")
	}

	return nil
}

func runList() error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	var manifestPath string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	absPath, err := resolveManifestPath(manifestPath)
	if err != nil {
		return err
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Tree.Children) == 0 {
		if jsonOutput {
			output := map[string]interface{}{
				"project":  m.Tree.Name,
				"features": []interface{}{},
			}
			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println("No children defined in manifest.")
			fmt.Println("Create one with: feat split \"\" <feature-name>")
		}
		return nil
	}

	if jsonOutput {
		data, err := formatter.FormatListJSON(m)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		printer := tree.NewPrinter()
		fmt.Print(printer.Print(m))
	}

	return nil
}

func runSplit() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: feat split <parent-path> <new-name>\n\nExamples:\n  feat split auth confirmation\n  feat split auth/password-reset email-template\n  feat split \"\" new-system")
	}

	parentPath := os.Args[2]
	newName := os.Args[3]

	fs := flag.NewFlagSet("split", flag.ContinueOnError)
	var manifestPath string
	var createFiles bool
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	fs.BoolVar(&createFiles, "create-files", true, "Create empty files on disk")
	if err := fs.Parse(os.Args[4:]); err != nil {
		return err
	}

	absPath, err := resolveManifestPath(manifestPath)
	if err != nil {
		return err
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	result, err := split.Split(m, split.Options{
		ParentPath:  parentPath,
		NewName:     newName,
		CreateFiles: createFiles,
		ManifestDir: filepath.Dir(absPath),
	})
	if err != nil {
		return err
	}

	if err := m.Save(absPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"path":     result.NewPath,
			"files":    result.FilesCreated,
			"manifest": absPath,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Print(split.FormatResult(result))
	}

	return nil
}

func runStatus() error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	var manifestPath string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	absPath, err := resolveManifestPath(manifestPath)
	if err != nil {
		return err
	}

	projectRoot := filepath.Dir(absPath)
	mgr := state.NewManager(projectRoot)

	s, err := mgr.GetCurrent()
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	// Load the feature if there's a current one
	var result *loader.Result
	if s != nil && s.FeaturePath != "" {
		m, err := manifest.Load(absPath)
		if err != nil {
			return fmt.Errorf("loading manifest: %w", err)
		}

		l := loader.New(m, absPath)
		result, err = l.Load(s.FeaturePath)
		if err != nil {
			// Don't fail if feature not found, just don't include files
			result = nil
		}
	}

	if jsonOutput {
		data, err := formatter.FormatStatusJSON(s, result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Print(state.FormatState(s))
		if result != nil {
			fmt.Println()
			fmt.Print(loader.FormatResult(result))
		}
	}

	return nil
}

func runValidate() error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	var manifestPath string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	absPath, err := resolveManifestPath(manifestPath)
	if err != nil {
		return err
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	issues := m.Validate()

	if jsonOutput {
		output := map[string]interface{}{
			"valid":  len(issues) == 0,
			"issues": issues,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(issues) == 0 {
			fmt.Println("✓ Manifest is valid")
		} else {
			fmt.Printf("Found %d issue(s):\n", len(issues))
			for _, issue := range issues {
				fmt.Printf("  - %s\n", issue)
			}
		}
	}

	return nil
}

func runWork() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: feat work <feature-path>\n\nExamples:\n  feat work auth/login\n  feat work payments/stripe-webhook")
	}

	featurePath := os.Args[2]

	fs := flag.NewFlagSet("work", flag.ContinueOnError)
	var manifestPath string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	if err := fs.Parse(os.Args[3:]); err != nil {
		return err
	}

	absPath, err := resolveManifestPath(manifestPath)
	if err != nil {
		return err
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	l := loader.New(m, absPath)
	result, err := l.Load(featurePath)
	if err != nil {
		return err
	}

	projectRoot := filepath.Dir(absPath)
	mgr := state.NewManager(projectRoot)
	if err := mgr.SetCurrent(featurePath, absPath); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"feature":   result.FeaturePath,
			"files":     result.Files,
			"tests":     result.Tests,
			"ancestors": result.AncestorFiles,
			"missing":   result.MissingFiles,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Print(loader.FormatResult(result))
	}

	return nil
}

func runParse() error {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	var manifestPath string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	absPath, err := resolveManifestPath(manifestPath)
	if err != nil {
		return err
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if jsonOutput {
		// Marshal the manifest directly to JSON
		type NodeJSON struct {
			Files    []string            `json:"files,omitempty"`
			Tests    []string            `json:"tests,omitempty"`
			Children map[string]NodeJSON `json:"children,omitempty"`
		}

		var convertNode func(n manifest.Node) NodeJSON
		convertNode = func(n manifest.Node) NodeJSON {
			result := NodeJSON{
				Files: n.Files,
				Tests: n.Tests,
			}
			if len(n.Children) > 0 {
				result.Children = make(map[string]NodeJSON)
				for name, child := range n.Children {
					result.Children[name] = convertNode(child)
				}
			}
			return result
		}

		output := map[string]interface{}{
			"project": m.Tree.Name,
			"files":   m.Tree.Files,
			"config": map[string]interface{}{
				"max_files": m.Config.GetMaxFiles(),
			},
			"children": convertNode(manifest.Node{Children: m.Tree.Children}),
		}

		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Manifest: %s\n", absPath)
		fmt.Printf("Project: %s\n", m.Tree.Name)
		if len(m.Tree.Files) > 0 {
			fmt.Printf("Root files: %v\n", m.Tree.Files)
		}
		fmt.Println()
		printManifest(m, 0)
	}

	return nil
}

// resolveManifestPath resolves and validates a manifest path.
func resolveManifestPath(manifestPath string) (string, error) {
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if we're in the right directory
			if _, err := os.Stat("feat.yaml"); err == nil && manifestPath == "feat.yaml" {
				// They ran without -f but there's a feat.yaml here
				abs, _ := filepath.Abs("feat.yaml")
				return abs, nil
			}
			return "", fmt.Errorf("manifest not found: %s\nRun 'feat init' to create one, or specify with -f", absPath)
		}
		return "", fmt.Errorf("checking manifest: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("manifest path is a directory: %s", absPath)
	}

	return absPath, nil
}

func printManifest(m *manifest.Manifest, indent int) {
	for name, node := range m.Tree.Children {
		printNode(name, node, indent)
	}
}

func printNode(name string, n manifest.Node, indent int) {
	prefix := strings.Repeat("  ", indent)

	if n.IsFeature() {
		fmt.Printf("%s%s (feature)\n", prefix, name)
		for _, file := range n.Files {
			fmt.Printf("%s  - %s\n", prefix, file)
		}
		if len(n.Tests) > 0 {
			fmt.Printf("%s  [tests: %v]\n", prefix, n.Tests)
		}
	} else {
		fmt.Printf("%s%s/ (boundary)\n", prefix, name)
		if len(n.Files) > 0 {
			fmt.Printf("%s  [files: %v]\n", prefix, n.Files)
		}
		if len(n.Tests) > 0 {
			fmt.Printf("%s  [tests: %v]\n", prefix, n.Tests)
		}
	}

	for childName, child := range n.Children {
		printNode(childName, child, indent+1)
	}
}
