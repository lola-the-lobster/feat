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
