package main

import "html/template"

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
	Breadcrumb []Breadcrumb
}

// Breadcrumb represents a navigation breadcrumb
type Breadcrumb struct {
	Title string
	URL   string
}

// Site represents the entire documentation site
type Site struct {
	Pages       []Page
	Sections    map[string][]Page
	NavTree     []NavItem
	Config      SiteConfig
	CurrentPage Page
}

// NavItem represents an item in the navigation tree
type NavItem struct {
	Title    string
	URL      string
	Children []NavItem
	Active   bool
}

// SiteConfig holds configuration from config.yaml
type SiteConfig struct {
	SiteName     string   `yaml:"siteName"`
	Description  string   `yaml:"description"`
	BaseURL      string   `yaml:"baseURL"`
	GithubRepo   string   `yaml:"githubRepo"`
	DefaultOrder []string `yaml:"defaultOrder"`
}
