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
	baseURL  = flag.String("base-url", "", "Base URL for assets (e.g., /myflow)")
)

// Define standard paths relative to repo root
const (
	docsDir      = "docs"         // Source markdown files
	publicDir    = "public"       // Output directory
	cmdDir       = "cmd/doc-generator"
	templatesDir = "templates"    // HTML templates (relative to cmdDir)
	staticDir    = "static"       // Static assets (relative to cmdDir)
	configFile   = "config.yaml"  // Site configuration (relative to cmdDir)
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

	// Verify repository structure
	cmdPath := filepath.Join(absRoot, cmdDir)
  verifyPaths(absRoot, cmdPath)

	// Load site configuration
	config, err := loadConfig(cmdPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set BaseURL from flag or default to empty
	config.BaseURL = strings.TrimRight(*baseURL, "/")

	// Generate the site
	if err := generateSite(absRoot, cmdPath, config); err != nil {
		log.Fatalf("Failed to generate site: %v", err)
	}

	if *watch {
		fmt.Printf("Watching for changes in %s... (press Ctrl+C to stop)\n", filepath.Join(absRoot, docsDir))
	}
}

func verifyPaths(absRoot, cmdPath string) map[string]string {
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
	return requiredPaths
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

	// Copy static assets to the correct location based on BaseURL
	staticPath := filepath.Join(cmdPath, staticDir)
	staticDestPath := outputDir
	if config.BaseURL != "" {
		staticDestPath = filepath.Join(outputDir, strings.TrimPrefix(config.BaseURL, "/"))
		if err := os.MkdirAll(staticDestPath, 0755); err != nil {
			return fmt.Errorf("failed to create base URL directory: %v", err)
		}
	}
	if err := copyDir(staticPath, staticDestPath); err != nil {
		return fmt.Errorf("failed to copy static assets: %v", err)
	}

	// Process markdown files
	docsPath := filepath.Join(rootDir, docsDir)
	allPages, err := processAllPages(docsPath, config)
	if err != nil {
		return err
	}

	// Build site structure
	site := buildSiteStructure(allPages, config)

	// Process templates and render pages
	templatesPath := filepath.Join(cmdPath, templatesDir)
	if err := renderPages(site, templatesPath, outputDir); err != nil {
		return fmt.Errorf("failed to render pages: %v", err)
	}

	return nil
}

func processAllPages(docsPath string, config SiteConfig) ([]Page, error) {
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
			// Adjust page URLPath based on BaseURL
			if config.BaseURL != "" {
				page.URLPath = filepath.Join(strings.TrimPrefix(config.BaseURL, "/"), page.URLPath)
			}
			allPages = append(allPages, page)
		}
		return nil
	})
	return allPages, err
}

func buildSiteStructure(allPages []Page, config SiteConfig) Site {
	// Build sections (original implementation)
	sections := make(map[string][]Page)
	for _, page := range allPages {
		if page.Section == "" {
			page.Section = "General"
		}
		sections[page.Section] = append(sections[page.Section], page)
	}

	// Build directory tree
	dirTree := buildDirTree(allPages)
	
	// Build tag map
	tagMap := make(map[string][]Page)
	for _, page := range allPages {
		for _, tag := range page.Tags {
			tagMap[tag] = append(tagMap[tag], page)
		}
	}

	// Create site with the new structure
	site := Site{
		Pages:    allPages,
		Sections: sections,
		Config:   config,
		NavTree:  buildNavTree(sections, dirTree, tagMap, config),
		DirTree:  dirTree,
		TagMap:   tagMap,
	}

	// Generate breadcrumbs
	for i := range site.Pages {
		site.Pages[i].Breadcrumb = generateBreadcrumbs(site.Pages[i], sections)
	}

	return site
}

// buildDirTree creates a hierarchical directory structure from pages
func buildDirTree(pages []Page) *Directory {
	root := &Directory{
		Name: "Home",
		Path: "",
	}
	
	dirMap := make(map[string]*Directory)
	dirMap[""] = root
	
	// First pass: create all directories
	for _, page := range pages {
		parts := strings.Split(page.DirPath, "/")
		current := ""
		parent := ""
		
		// Create each directory in the path if it doesn't exist
		for _, part := range parts {
			if part == "" {
				continue
			}
			
			parent = current
			if current != "" {
				current = current + "/" + part
			} else {
				current = part
			}
			
			if _, exists := dirMap[current]; !exists {
				dir := &Directory{
					Name: part,
					Path: current,
				}
				
				// Link to parent
				if parentDir, ok := dirMap[parent]; ok {
					dir.ParentDir = parentDir
					parentDir.SubDirs = append(parentDir.SubDirs, dir)
				}
				
				dirMap[current] = dir
			}
		}
	}
	
	// Second pass: add pages to their directories
	for _, page := range pages {
		if dir, ok := dirMap[page.DirPath]; ok {
			dir.Pages = append(dir.Pages, page)
		}
	}
	
	return root
}

func renderPages(site Site, templatesPath, outputDir string) error {
	tmpl := template.New("").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
		"contains": strings.Contains,
		"add": func(a, b int) int { return a + b },
		"now": time.Now,
		"assetPath": func(path string) string {
			return site.AssetPath(path)
		},
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
		if err := generateIndexPage(site, tmpl, outputDir); err != nil {
			return err
		}
	}

	return nil
}

func generateIndexPage(site Site, tmpl *template.Template, outputDir string) error {
	indexPage := findIndexPage(site.Pages)
	site.CurrentPage = indexPage
	updateActiveNav(&site.NavTree, indexPage.URLPath)

	// Create index.html at the base URL location
	indexPath := filepath.Join(outputDir, "index.html")
	if site.Config.BaseURL != "" {
		indexPath = filepath.Join(outputDir, strings.TrimPrefix(site.Config.BaseURL, "/"), "index.html")
	}

	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %v", err)
	}

	file, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("failed to create index.html: %v", err)
	}
	defer file.Close()

	if err := tmpl.ExecuteTemplate(file, "base", site); err != nil {
		return fmt.Errorf("failed to render index.html: %v", err)
	}

	// If we have a base URL, create a redirect at the root
	if site.Config.BaseURL != "" {
		rootIndex := filepath.Join(outputDir, "index.html")
		redirectContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0;url=%s/">
</head>
</html>`, site.Config.BaseURL)
		if err := os.WriteFile(rootIndex, []byte(redirectContent), 0644); err != nil {
			return fmt.Errorf("failed to create root redirect: %v", err)
		}
	}

	return nil
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

	// Command line flag takes precedence over config file
	if *baseURL != "" {
		config.BaseURL = strings.TrimRight(*baseURL, "/")
	}

	return config, nil
}
