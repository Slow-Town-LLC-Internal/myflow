# MyFlow Doc Generator Commands & Style Guide

## Build & Run Commands
- **Build:** `go build`
- **Run:** `./doc-generator [flags]`
- **Run with watch mode:** `./doc-generator -watch`
- **Set custom repo root:** `./doc-generator -root /path/to/repo`
- **Set custom port:** `./doc-generator -port 8080` 
- **Set base URL:** `./doc-generator -base-url /myflow`
- **Run tests:** `go test ./...`
- **Run single test:** `go test -run TestName`

## Code Style Guidelines
- **Imports:** Group standard library imports first, then external packages
- **Error Handling:** Check errors immediately, use descriptive error messages with fmt.Errorf
- **Function Size:** Keep functions focused and under 50 lines
- **Path Handling:** Use filepath for cross-platform path manipulation
- **Naming:** Use camelCase for variables, PascalCase for exported functions/types
- **Comments:** Document all exported functions, types, and constants
- **File Organization:** Group related functionality in separate files
- **Config:** Use YAML for configuration with typed structs
- **HTML Templates:** Use Go's html/template package with components