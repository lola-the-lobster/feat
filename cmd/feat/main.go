package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lola-the-lobster/feat/internal/manifest"
	"github.com/lola-the-lobster/feat/internal/tree"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: feat <command> [args]")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  list     Show feature tree")
		fmt.Fprintln(os.Stderr, "  parse    Parse .feat.yml and dump structure")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func runList() error {
	var manifestPath string
	flag.StringVar(&manifestPath, "f", ".feat.yml", "Path to manifest file")
	flag.CommandLine.Parse(os.Args[2:])

	// Resolve absolute path
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest not found: %s", absPath)
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return err
	}

	printer := tree.NewPrinter()
	fmt.Print(printer.Print(m))

	return nil
}

func runParse() error {
	var manifestPath string
	flag.StringVar(&manifestPath, "f", ".feat.yml", "Path to manifest file")
	flag.CommandLine.Parse(os.Args[2:])

	// Resolve absolute path
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest not found: %s", absPath)
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return err
	}

	fmt.Printf("Manifest: %s\n", absPath)
	fmt.Println()
	printManifest(m, 0)

	return nil
}

func printManifest(m *manifest.Manifest, indent int) {
	for name, feature := range m.Features {
		printFeature(name, feature, indent)
	}
}

func printFeature(name string, f manifest.Feature, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	if f.IsLeaf() {
		fmt.Printf("%s%s (feature)\n", prefix, name)
		for _, file := range f.Files {
			fmt.Printf("%s  - %s\n", prefix, file)
		}
	} else {
		fmt.Printf("%s%s/\n", prefix, name)
		if f.Interface != "" {
			fmt.Printf("%s  [interface: %s]\n", prefix, f.Interface)
		}
		if len(f.Deps) > 0 {
			fmt.Printf("%s  [deps: %v]\n", prefix, f.Deps)
		}
	}

	// Print children
	for childName, child := range f.Children {
		printFeature(childName, child, indent+1)
	}
}
