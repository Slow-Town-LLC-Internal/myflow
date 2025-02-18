package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
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
	SiteName     string `yaml:"siteName"`
	Description  string `yaml:"description"`
	BaseURL      string `yaml:"baseURL"`
	GithubRepo   string `yaml:"githubRepo"`
	DefaultOrder []string `yaml:"defaultOrder"`
}

func main() {
	// 1. Load site configuration
	configData, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config SiteConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// 2. Create output directory structure
	os.RemoveAll("public")
	os.MkdirAll("public", 0755)

	// 3. Copy static assets
	if err := copyDir("static", "public"); err != nil {
		log.Fatalf("Failed to copy static assets: %v", err)
	}

	// 4. Process all markdown files
	var allPages []Page
	err = filepath.Walk("content", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".md") {
			page, err := processMarkdown(path, config)
			if err != nil {
				log.Printf("Warning: Failed to process %s: %v", path, err)
				return nil
			}
			allPages = append(allPages, page)
		} else if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			// Copy non-markdown files (like images)
			relPath, _ := filepath.Rel("content", path)
			destPath := filepath.Join("public", relPath)
			os.MkdirAll(filepath.Dir(destPath), 0755)
			if err := copyFile(path, destPath); err != nil {
				log.Printf("Warning: Failed to copy %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to walk content directory: %v", err)
	}

	// 5. Organize pages by section
	sections := make(map[string][]Page)
	for _, page := range allPages {
		if page.Section == "" {
			page.Section = "General"
		}
		sections[page.Section] = append(sections[page.Section], page)
	}

	// Sort pages by Order within sections
	for section := range sections {
		sort.Slice(sections[section], func(i, j int) bool {
			return sections[section][i].Order < sections[section][j].Order
		})
	}

	// 6. Build navigation tree
	navTree := buildNavTree(sections, config)

	// 7. Generate breadcrumbs for each page
	for i := range allPages {
		allPages[i].Breadcrumb = generateBreadcrumbs(allPages[i], sections)
	}

	// 8. Render each page with templates
	tmpl := template.New("").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
		"contains": strings.Contains,
		"add": func(a, b int) int {
			return a + b
		},
		"now": time.Now,
	})

	tmpl, err = tmpl.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	tmpl, err = tmpl.ParseGlob("templates/components/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse component templates: %v", err)
	}

	site := Site{
		Pages:    allPages,
		Sections: sections,
		Config:   config,
		NavTree:  navTree,
	}

	for _, page := range allPages {
		site.CurrentPage = page

		// Update Active state in navigation
		updateActiveNav(&site.NavTree, page.URLPath)

		outPath := filepath.Join("public", page.URLPath)
		os.MkdirAll(filepath.Dir(outPath), 0755)

		file, err := os.Create(outPath)
		if err != nil {
			log.Fatalf("Failed to create output file %s: %v", outPath, err)
		}

		if err := tmpl.ExecuteTemplate(file, "base", site); err != nil {
			file.Close()
			log.Fatalf("Failed to render template for %s: %v", outPath, err)
		}
		file.Close()
	}

	// 9. Generate index.html
	if len(allPages) > 0 {
		// Find the README or index page, or use the first page
		var indexPage Page
		for _, page := range allPages {
			if strings.HasSuffix(page.RelPath, "README.md") ||
			   strings.HasSuffix(page.RelPath, "index.md") {
				indexPage = page
				break
			}
		}

		if indexPage.RelPath == "" {
			// No index found, use first page
			indexPage = allPages[0]
		}

		site.CurrentPage = indexPage
		updateActiveNav(&site.NavTree, indexPage.URLPath)

		indexFile, err := os.Create("public/index.html")
		if err != nil {
			log.Fatalf("Failed to create index.html: %v", err)
		}

		if err := tmpl.ExecuteTemplate(indexFile, "base", site); err != nil {
			indexFile.Close()
			log.Fatalf("Failed to render index.html: %v", err)
		}
		indexFile.Close()
	}

	// 10. Generate 404 page
	notFoundFile, err := os.Create("public/404.html")
	if err != nil {
		log.Fatalf("Failed to create 404.html: %v", err)
	}

	if err := tmpl.ExecuteTemplate(notFoundFile, "404", site); err != nil {
		notFoundFile.Close()
		log.Fatalf("Failed to render 404.html: %v", err)
	}
	notFoundFile.Close()

	fmt.Println("Site generated successfully in the 'public' directory!")
}

func processMarkdown(filePath string, config SiteConfig) (Page, error) {
	content, err := ioutil.ReadFile(filePath)
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
	relPath, err := filepath.Rel("content", filePath)
	if err != nil {
		return Page{}, err
	}

	urlPath := strings.TrimSuffix(relPath, ".md") + ".html"
	// Handle README files
	if strings.HasSuffix(urlPath, "README.html") {
		urlPath = strings.TrimSuffix(urlPath, "README.html") + "index.html"
	}

	return Page{
		FrontMatter: frontMatter,
		Content:     template.HTML(buf.String()),
		RelPath:     relPath,
		URLPath:     urlPath,
	}, nil
}

func buildNavTree(sections map[string][]Page, config SiteConfig) []NavItem {
	var navTree []NavItem

	// Sort sections according to config.DefaultOrder
	sectionNames := make([]string, 0, len(sections))
	for section := range sections {
		sectionNames = append(sectionNames, section)
	}

	sort.Slice(sectionNames, func(i, j int) bool {
		// If section is in DefaultOrder, use its position
		iPos := -1
		jPos := -1
		for idx, name := range config.DefaultOrder {
			if name == sectionNames[i] {
				iPos = idx
			}
			if name == sectionNames[j] {
				jPos = idx
			}
		}

		// If both sections are in DefaultOrder, use that order
		if iPos >= 0 && jPos >= 0 {
			return iPos < jPos
		}

		// If only one section is in DefaultOrder, prioritize it
		if iPos >= 0 {
			return true
		}
		if jPos >= 0 {
			return false
		}

		// Otherwise, sort alphabetically
		return sectionNames[i] < sectionNames[j]
	})

	for _, section := range sectionNames {
		pages := sections[section]

		var children []NavItem
		for _, page := range pages {
			children = append(children, NavItem{
				Title: page.Title,
				URL:   "/" + page.URLPath,
			})
		}

		navTree = append(navTree, NavItem{
			Title:    section,
			Children: children,
		})
	}

	return navTree
}

func generateBreadcrumbs(page Page, sections map[string][]Page) []Breadcrumb {
	breadcrumbs := []Breadcrumb{
		{Title: "Home", URL: "/"},
	}

	if page.Section != "" && page.Section != "General" {
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			Title: page.Section,
			URL:   "#", // We don't have section index pages
		})
	}

	breadcrumbs = append(breadcrumbs, Breadcrumb{
		Title: page.Title,
		URL:   "/" + page.URLPath,
	})

	return breadcrumbs
}

func updateActiveNav(navTree *[]NavItem, currentPath string) {
	for i := range *navTree {
		(*navTree)[i].Active = false
		for j := range (*navTree)[i].Children {
			// Check if this nav item or any child matches current path
			isActive := ("/" + currentPath) == (*navTree)[i].Children[j].URL
			(*navTree)[i].Children[j].Active = isActive
			if isActive {
				(*navTree)[i].Active = true
			}
		}
	}
}

func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}
