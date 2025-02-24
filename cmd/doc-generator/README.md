# MyFlow Documentation Generator

A static site generator that converts markdown files into a structured HTML documentation site.

## Features

- Markdown to HTML conversion with frontmatter support
- Automatic navigation with sections and breadcrumbs
- Directory-based navigation to browse by folder structure
- Tag-based content filtering
- Code syntax highlighting with Prism.js
- Responsive design
- GitHub edit links
- Google Analytics integration

## Usage

```bash
# Basic usage
./doc-generator

# With custom options
./doc-generator -root /path/to/repo -port 8080 -base-url /myflow -watch
```

## Configuration

Create a `config.yaml` file in the project directory:

```yaml
siteName: "Project Documentation"
description: "Documentation for your project"
baseURL: ""
githubRepo: "username/repo"
GATrackingID: "G-XXXXXXXXXX"
defaultOrder:
  - "Getting Started"
  - "Guides"
  - "Reference"
```

## Markdown Structure

Place markdown files in the `docs/` directory. Add frontmatter for metadata:

```markdown
---
title: "Page Title"
description: "Page description"
section: "Section Name"
order: 1
tags: ["tag1", "tag2"]
---

# Content starts here
```

## Development Status

Current implementation includes:
- ✅ Google Analytics integration
- ✅ Directory-based navigation structure
- ✅ Tag UI and filtering infrastructure
- ⚠️ **TODO**: Fix tag data propagation from markdown frontmatter to navigation items

The tag functionality is partially implemented. Tags defined in frontmatter aren't correctly passed to the navigation items yet. To complete this:
1. Verify tag extraction from frontmatter
2. Check how tags are passed through the page processing pipeline
3. Debug tag attachment to navigation items in navigation.go