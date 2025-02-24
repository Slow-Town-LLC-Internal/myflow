package main

import (
	"sort"
	"strings"
)

// buildNavTree creates a navigation tree from directories, tags, and sections
func buildNavTree(sections map[string][]Page, dirTree *Directory, tagMap map[string][]Page, config SiteConfig) []NavItem {
	var navTree []NavItem
	
	// 1. Add Directories section
	dirNavItem := NavItem{
		Title:  "Directories",
		IsDir:  true,
		DirPath: "",
	}
	
	// Add root directory content
	var rootPages []NavItem
	for _, page := range dirTree.Pages {
		rootPages = append(rootPages, NavItem{
			Title: page.Title,
			URL:   "/" + page.URLPath,
			Tags:  page.Tags,
		})
	}
	
	// Add subdirectories
	var subDirs []NavItem
	for _, subDir := range dirTree.SubDirs {
		subDirs = append(subDirs, NavItem{
			Title:   subDir.Name,
			URL:     "#",
			IsDir:   true,
			DirPath: subDir.Path,
		})
	}
	
	// Sort subdirectories alphabetically
	sort.Slice(subDirs, func(i, j int) bool {
		return subDirs[i].Title < subDirs[j].Title
	})
	
	// Combine root pages and subdirectories
	dirNavItem.Children = append(rootPages, subDirs...)
	navTree = append(navTree, dirNavItem)
	
	// 2. Add Tags section
	if len(tagMap) > 0 {
		tagNavItem := NavItem{
			Title: "Tags",
		}
		
		// Get all tags and sort them
		tags := make([]string, 0, len(tagMap))
		for tag := range tagMap {
			tags = append(tags, tag)
		}
		sort.Strings(tags)
		
		// Create child items for each tag
		for _, tag := range tags {
			tagNavItem.Children = append(tagNavItem.Children, NavItem{
				Title: tag,
				URL:   "#tag=" + tag,
				IsTag: true,
				Tag:   tag,
			})
		}
		
		navTree = append(navTree, tagNavItem)
	}
	
	// 3. Add Sections (original implementation)
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
				Tags:  page.Tags,
			})
		}

		navTree = append(navTree, NavItem{
			Title:    section,
			Children: children,
		})
	}

	return navTree
}

// generateBreadcrumbs creates breadcrumbs for a page
func generateBreadcrumbs(page Page, sections map[string][]Page) []Breadcrumb {
	breadcrumbs := []Breadcrumb{
		{Title: "Home", URL: "/"},
	}

	if page.Section != "" && page.Section != "General" {
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			Title: page.Section,
			URL:   "#",
		})
	}

	breadcrumbs = append(breadcrumbs, Breadcrumb{
		Title: page.Title,
		URL:   "/" + page.URLPath,
	})

	return breadcrumbs
}

// updateActiveNav updates active state in navigation tree
func updateActiveNav(navTree *[]NavItem, currentPath string) {
	for i := range *navTree {
		(*navTree)[i].Active = false
		
		// Handle special cases for directories and tags
		if (*navTree)[i].IsDir && currentPath != "" {
			// For directory navigation, set active if current path starts with this directory path
			for j := range (*navTree)[i].Children {
				if (*navTree)[i].Children[j].IsDir {
					// For subdirectories, check if the current path starts with this directory path
					dirPath := (*navTree)[i].Children[j].DirPath
					isActive := dirPath != "" && strings.HasPrefix(currentPath, dirPath)
					(*navTree)[i].Children[j].Active = isActive
					if isActive {
						(*navTree)[i].Active = true
					}
				} else {
					// For pages, check exact match
					isActive := ("/" + currentPath) == (*navTree)[i].Children[j].URL
					(*navTree)[i].Children[j].Active = isActive
					if isActive {
						(*navTree)[i].Active = true
					}
				}
			}
		} else {
			// Regular section navigation (unchanged)
			for j := range (*navTree)[i].Children {
				isActive := ("/" + currentPath) == (*navTree)[i].Children[j].URL
				(*navTree)[i].Children[j].Active = isActive
				if isActive {
					(*navTree)[i].Active = true
				}
			}
		}
	}
}
