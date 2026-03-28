# feat

Feature-centric context management for agentic coding.

## What's New: The feat.yaml Schema

The manifest file defines your project's feature hierarchy:

```yaml
config:
  max_files: 3                 # Max files per feature + ancestors
  workflow: [scaffold, fix, build, test, done]
tree:
  name: my-project
  files: [go.mod, README.md]   # Root-level files
  children:
    auth:                      # Boundary (has children)
      files: [auth/interface.go]
      children:
        login:                 # Feature (no children, has files/tests)
          files: [auth/login.go]
          tests: [auth/login_test.go]
```

**Key fields:**
- `config.max_files` — Maximum files to load (feature + ancestor files)
- `config.workflow` — Custom workflow steps for features
- `tree.name` — Project name
- `tree.files` — Shared files at root level
- `tree.children` — Feature hierarchy (boundaries nest, features are leaves)
- Node fields: `files` (implementation), `tests` (test files), `children` (sub-features)

## Overview

`feat` organizes code by feature, not by layer. Instead of loading entire packages, you work on specific features with their relevant files and ancestor context.



## Installation

### Requirements

- Go 1.21 or later

### From Source

Clone the repository and build:

```bash
git clone https://github.com/plor/feat.git
cd feat
go build -o feat ./cmd/feat
```

Or install directly with `go install`:

```bash
go install github.com/plor/feat/cmd/feat@latest
```

### Verify Installation

```bash
feat --version
```

## Quick Start Tutorial

Create a new project and initialize it with feat:

```bash
mkdir myproject && cd myproject
```

Initialize feat in your project:

```bash
feat init
```

This creates:
- `.feat.yml` — The manifest file defining your feature hierarchy
- `.feat/` — Directory containing state and metadata

### Generated .feat.yml

After running `feat init`, you'll have a basic manifest:

```yaml
config:
  max_files: 3
  workflow: [scaffold, fix, build, test, done]
tree:
  name: myproject
  files: []
  children: {}
```

### Initial .feat/state.json

The state file tracks your current context:

```json
{
  "current_feature": "",
  "workflow_state": ""
}
```

### View Available Features

List all features in your project:

```bash
feat list
```

### Start Working on a Feature

Begin working on a specific feature:

```bash
feat work <feature-id>
```

This loads the feature's context, including its files and ancestor context.


## Commands

- `feat init` — Create a new feat.yaml manifest
- `feat list` — Show feature tree
- `feat work <feature>` — Load a feature's context
- `feat split <parent> <name>` — Create a new feature
- `feat status` — Show current feature context
- `feat transition <step>` — Update feature workflow state
- `feat validate` — Check manifest for issues

## Example

```bash
# Initialize a project
feat init --name my-app

# Create a feature
feat split "" auth
feat split auth login

# Work on a feature
feat work auth/login

# Check status
feat status

# Move to next workflow step
feat transition build
```

## License

MIT
