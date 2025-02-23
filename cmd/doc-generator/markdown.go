package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

// processMarkdown processes a markdown file and returns a Page
func processMarkdown(filePath, baseDir string, config SiteConfig) (Page, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Page{}, err
	}

	// Check if the file has frontmatter
	var frontMatter FrontMatter
	var markdown []byte

	if bytes.HasPrefix(content, []byte("---")) {
		parts := bytes.SplitN(content, []byte("---"), 3)
		if len(parts) >= 3 {
			if err := yaml.Unmarshal(parts[1], &frontMatter); err != nil {
				return Page{}, fmt.Errorf("invalid frontmatter: %v", err)
			}
			markdown = parts[2]
		} else {
			markdown = content
		}
	} else {
		markdown = content
	}

	// If no title in frontmatter, use filename
	if frontMatter.Title == "" {
		frontMatter.Title = strings.TrimSuffix(filepath.Base(filePath), ".md")
	}

	// Convert markdown to HTML
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // Allows HTML in markdown
		),
	)

	if err := md.Convert(markdown, &buf); err != nil {
		return Page{}, err
	}

	// Generate relative and URL paths
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return Page{}, err
	}

	urlPath := strings.TrimSuffix(relPath, ".md") + ".html"
	// Handle README files
	if strings.HasSuffix(urlPath, "README.html") {
		urlPath = strings.TrimSuffix(urlPath, "README.html") + "index.html"
	}

	// Ensure URL paths use forward slashes
	urlPath = filepath.ToSlash(urlPath)

	// Prepend baseURL if configured
	if config.BaseURL != "" {
		urlPath = filepath.Join(config.BaseURL, urlPath)
		urlPath = filepath.ToSlash(urlPath) // Convert to forward slashes again
	}

	return Page{
		FrontMatter: frontMatter,
		Content:     template.HTML(buf.String()),
		RelPath:     relPath,
		URLPath:     urlPath,
	}, nil
}
