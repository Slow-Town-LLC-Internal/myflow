package main

import (
	"html/template"
	"path/filepath"
)

// FrontMatter represents the YAML metadata at the top of markdown files
type FrontMatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Section     string   `yaml:"section"`
	Order       int      `yaml:"order"`
	Tags        []string `yaml:"tags"`
}

// Page represents a processed markdown page
type Page struct {
	FrontMatter
	Content    template.HTML
	RelPath    string
	URLPath    string
	DirPath    string // Directory path relative to docs root
	Breadcrumb []Breadcrumb
}

// Breadcrumb represents a navigation breadcrumb
type Breadcrumb struct {
	Title string
	URL   string
}

// Directory represents a directory in the docs structure
type Directory struct {
	Name     string
	Path     string
	Pages    []Page
	SubDirs  []*Directory
	ParentDir *Directory
}

// Site represents the entire documentation site
type Site struct {
	Pages       []Page
	Sections    map[string][]Page
	NavTree     []NavItem
	DirTree     *Directory  // Root directory of the documentation
	TagMap      map[string][]Page // Map of tags to pages
	Config      SiteConfig
	CurrentPage Page
	CurrentDir  string  // Current directory path relative to docs root
}

// AssetPath returns the correct path for assets considering the base URL
func (s Site) AssetPath(path string) string {
	if s.Config.BaseURL == "" {
		return path
	}
	return filepath.Join(s.Config.BaseURL, path)
}

// NavItem represents an item in the navigation tree
type NavItem struct {
	Title      string
	URL        string
	Children   []NavItem
	Active     bool
	IsDir      bool      // Whether this item is a directory
	IsTag      bool      // Whether this item is a tag
	DirPath    string    // Directory path if this is a directory
	Tag        string    // Tag name if this is a tag
	Tags       []string  // Tags for this page
}

// SiteConfig holds configuration from config.yaml
type SiteConfig struct {
	SiteName     string   `yaml:"siteName"`
	Description  string   `yaml:"description"`
	BaseURL      string   `yaml:"baseURL"`
	GithubRepo   string   `yaml:"githubRepo"`
	DefaultOrder []string `yaml:"defaultOrder"`
	GATrackingID string   `yaml:"GATrackingID"`
}
