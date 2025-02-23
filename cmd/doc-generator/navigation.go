package main

import "sort"

// buildNavTree creates a navigation tree from sections
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
		for j := range (*navTree)[i].Children {
			isActive := ("/" + currentPath) == (*navTree)[i].Children[j].URL
			(*navTree)[i].Children[j].Active = isActive
			if isActive {
				(*navTree)[i].Active = true
			}
		}
	}
}
