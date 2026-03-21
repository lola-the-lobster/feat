package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Handle version flag
	if os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "version" {
		fmt.Printf("feat version %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	// Handle help flag
	if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := runList(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "parse":
		if err := runParse(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "split":
		if err := runSplit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := runStatus(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "validate":
		if err := runValidate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "work":
		if err := runWork(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintln(os.Stderr, "Run 'feat help' for usage.")
		os.Exit(1)
	}
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
	fmt.Println("  -f <path>         Path to manifest file (default: feat.yaml)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  feat init                    # Create new manifest")
	fmt.Println("  feat list                    # Show all features")
	fmt.Println("  feat work auth/login         # Work on auth/login feature")
	fmt.Println("  feat split auth login-v2     # Create auth/login-v2 feature")
	fmt.Println("  feat status                  # Show current feature")
	fmt.Println("  feat validate                # Check for issues")
}

func runInit() error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	var manifestPath string
	fs.StringVar(&manifestPath, "f", "feat.yaml", "Path to manifest file")
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

	if err := manifest.Init(absPath); err != nil {
		return fmt.Errorf("creating manifest: %w", err)
	}

	fmt.Printf("Created manifest: %s\n", absPath)
	fmt.Println("Add features to get started:")
	fmt.Println("  feat split \"\" my-feature")

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

	if len(m.Features) == 0 {
		fmt.Println("No features defined in manifest.")
		fmt.Println("Create one with: feat split \"\" <feature-name>")
		return nil
	}

	printer := tree.NewPrinter()
	fmt.Print(printer.Print(m))

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

	fmt.Print(split.FormatResult(result))

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

	fmt.Print(state.FormatState(s))

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
	if len(issues) == 0 {
		fmt.Println("✓ Manifest is valid")
		return nil
	}

	fmt.Printf("Found %d issue(s):\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("  - %s\n", issue)
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

	fmt.Print(loader.FormatResult(result))

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

	fmt.Printf("Manifest: %s\n", absPath)
	fmt.Println()
	printManifest(m, 0)

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
	for name, feature := range m.Features {
		printFeature(name, feature, indent)
	}
}

func printFeature(name string, f manifest.Feature, indent int) {
	prefix := strings.Repeat("  ", indent)

	if f.IsLeaf() {
		fmt.Printf("%s%s (feature)\n", prefix, name)
		for _, file := range f.Files {
			fmt.Printf("%s  - %s\n", prefix, file)
		}
	} else {
		fmt.Printf("%s%s/\n", prefix, name)
		if len(f.Files) > 0 {
			fmt.Printf("%s  [files: %v]\n", prefix, f.Files)
		}
	}

	for childName, child := range f.Children {
		printFeature(childName, child, indent+1)
	}
}
