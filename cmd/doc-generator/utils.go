package main

import (
	"io"
	"os"
	"path/filepath"
)

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
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

// copyFile copies a single file
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

// findIndexPage finds the index page from a list of pages
func findIndexPage(pages []Page) Page {
	// First, look for README.md or index.md
	for _, page := range pages {
		if filepath.Base(page.RelPath) == "README.md" ||
			filepath.Base(page.RelPath) == "index.md" {
			return page
		}
	}

	// If no index found, use the first page
	if len(pages) > 0 {
		return pages[0]
	}

	// Return empty page if no pages exist
	return Page{}
}
