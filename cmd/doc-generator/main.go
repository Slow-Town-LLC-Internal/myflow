package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	repoRoot = flag.String("root", ".", "Repository root directory")
	watch    = flag.Bool("watch", false, "Watch for file changes and rebuild")
	port     = flag.String("port", "8080", "Port to serve the site")
)

// Define standard paths relative to repo root
const (
	docsDir      = "docs"      // Source markdown files
	publicDir    = "public"    // Output directory
	cmdDir       = "cmd/doc-generator"
	templatesDir = "templates" // HTML templates (relative to cmdDir)
	staticDir    = "static"    // Static assets (relative to cmdDir)
	configFile   = "config.yaml" // Site configuration (relative to cmdDir)
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flag.Parse()

	// Resolve absolute paths
	absRoot, err := filepath.Abs(*repoRoot)
	if err != nil {
		log.Fatalf("Failed to resolve repository root path: %v", err)
	}

	// Get the cmd/doc-generator directory
	cmdPath := filepath.Join(absRoot, cmdDir)

	// Verify repository structure
	requiredPaths := map[string]string{
		"docs":      filepath.Join(absRoot, docsDir),
		"templates": filepath.Join(cmdPath, templatesDir),
		"static":    filepath.Join(cmdPath, staticDir),
		"config":    filepath.Join(cmdPath, configFile),
	}

	for name, path := range requiredPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatalf("Required %s directory/file not found: %s", name, path)
		}
	}

	// Load site configuration
	config, err := loadConfig(cmdPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Generate the site
	if err := generateSite(absRoot, cmdPath, config); err != nil {
		log.Fatalf("Failed to generate site: %v", err)
	}

	if *watch {
		fmt.Printf("Watching for changes in %s... (press Ctrl+C to stop)\n", filepath.Join(absRoot, docsDir))
		// TODO: Implement file watching
	}
}

func loadConfig(cmdPath string) (SiteConfig, error) {
	configPath := filepath.Join(cmdPath, configFile)
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return SiteConfig{}, fmt.Errorf("failed to read config file: %v", err)
	}

	var config SiteConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return SiteConfig{}, fmt.Errorf("failed to parse config: %v", err)
	}

	return config, nil
}

func generateSite(rootDir, cmdPath string, config SiteConfig) error {
	// Setup output directory
	outputDir := filepath.Join(rootDir, publicDir)
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("failed to clean output directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Copy static assets
	staticPath := filepath.Join(cmdPath, staticDir)
	if err := copyDir(staticPath, outputDir); err != nil {
		return fmt.Errorf("failed to copy static assets: %v", err)
	}

	// Process markdown files
	docsPath := filepath.Join(rootDir, docsDir)
	var allPages []Page
	err := filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".md") {
			page, err := processMarkdown(path, docsPath, config)
			if err != nil {
				log.Printf("Warning: Failed to process %s: %v", path, err)
				return nil
			}
			allPages = append(allPages, page)
		} else if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			// Copy non-markdown files
			relPath, _ := filepath.Rel(docsPath, path)
			destPath := filepath.Join(outputDir, relPath)
			os.MkdirAll(filepath.Dir(destPath), 0755)
			if err := copyFile(path, destPath); err != nil {
				log.Printf("Warning: Failed to copy %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk docs directory: %v", err)
	}

	// Organize pages by section
	sections := make(map[string][]Page)
	for _, page := range allPages {
		if page.Section == "" {
			page.Section = "General"
		}
		sections[page.Section] = append(sections[page.Section], page)
	}

	// Build site structure
	site := Site{
		Pages:    allPages,
		Sections: sections,
		Config:   config,
		NavTree:  buildNavTree(sections, config),
	}

	// Generate breadcrumbs
	for i := range site.Pages {
		site.Pages[i].Breadcrumb = generateBreadcrumbs(site.Pages[i], sections)
	}

	// Process templates and render pages
	templatesPath := filepath.Join(cmdPath, templatesDir)
	if err := renderPages(site, templatesPath, outputDir); err != nil {
		return fmt.Errorf("failed to render pages: %v", err)
	}

	return nil
}

func renderPages(site Site, templatesPath, outputDir string) error {
	tmpl := template.New("").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
		"contains":  strings.Contains,
		"add": func(a, b int) int {
			return a + b
		},
		"now": time.Now,
	})

	// Parse templates
	tmpl, err := tmpl.ParseGlob(filepath.Join(templatesPath, "*.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to parse templates: %v", err)
	}

	tmpl, err = tmpl.ParseGlob(filepath.Join(templatesPath, "components/*.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to parse component templates: %v", err)
	}

	// Render each page
	for _, page := range site.Pages {
		site.CurrentPage = page
		updateActiveNav(&site.NavTree, page.URLPath)

		outPath := filepath.Join(outputDir, page.URLPath)
		os.MkdirAll(filepath.Dir(outPath), 0755)

		file, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %v", outPath, err)
		}

		if err := tmpl.ExecuteTemplate(file, "base", site); err != nil {
			file.Close()
			return fmt.Errorf("failed to render template for %s: %v", outPath, err)
		}
		file.Close()
	}

	// Generate index.html if needed
	if len(site.Pages) > 0 {
		indexPage := findIndexPage(site.Pages)
		site.CurrentPage = indexPage
		updateActiveNav(&site.NavTree, indexPage.URLPath)

		indexPath := filepath.Join(outputDir, "index.html")
		file, err := os.Create(indexPath)
		if err != nil {
			return fmt.Errorf("failed to create index.html: %v", err)
		}
		defer file.Close()

		if err := tmpl.ExecuteTemplate(file, "base", site); err != nil {
			return fmt.Errorf("failed to render index.html: %v", err)
		}
	}

	return nil
}
