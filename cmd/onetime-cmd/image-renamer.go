package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	// Command line arguments
	sourceDir := flag.String("dir", "vsco_photos", "Directory containing images to rename")
	prefix := flag.String("prefix", "vsco", "Prefix for renamed files")
	keepOriginals := flag.Bool("keep", false, "Keep original files (otherwise they will be replaced)")
	outputDir := flag.String("output", "", "Output directory for renamed files (default: same as source)")
	flag.Parse()

	// If output directory is not specified, use source directory
	targetDir := *sourceDir
	if *outputDir != "" {
		targetDir = *outputDir
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return
		}
	}

	// Get all image files in the directory
	imageFiles, err := getImageFiles(*sourceDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	if len(imageFiles) == 0 {
		fmt.Printf("No image files found in %s\n", *sourceDir)
		return
	}

	fmt.Printf("Found %d image files\n", len(imageFiles))

	// Sort files by modification time (oldest first)
	// This is a reasonable approximation for chronological order
	sort.Slice(imageFiles, func(i, j int) bool {
		fileI := filepath.Join(*sourceDir, imageFiles[i])
		fileJ := filepath.Join(*sourceDir, imageFiles[j])
		
		infoI, errI := os.Stat(fileI)
		infoJ, errJ := os.Stat(fileJ)
		
		if errI != nil || errJ != nil {
			return imageFiles[i] < imageFiles[j] // Fallback to alphabetical
		}
		
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Get current date for timestamp
	timestamp := time.Now().Format("20060102")

	// Process each file
	for i, file := range imageFiles {
		// Get file extension
		ext := strings.ToLower(filepath.Ext(file))
		
		// Generate new filename
		newFilename := fmt.Sprintf("%s-%s-%03d%s", *prefix, timestamp, i+1, ext)
		
		srcPath := filepath.Join(*sourceDir, file)
		destPath := filepath.Join(targetDir, newFilename)
		
		// Skip if destination file already exists
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("Skipping %s (destination file already exists)\n", file)
			continue
		}
		
		// Copy or rename the file
		if *keepOriginals || *outputDir != "" {
			// Copy file
			if err := copyFile(srcPath, destPath); err != nil {
				fmt.Printf("Error copying %s: %v\n", file, err)
				continue
			}
			fmt.Printf("Copied %s to %s\n", file, newFilename)
		} else {
			// Rename file
			if err := os.Rename(srcPath, destPath); err != nil {
				fmt.Printf("Error renaming %s: %v\n", file, err)
				continue
			}
			fmt.Printf("Renamed %s to %s\n", file, newFilename)
		}
	}
	
	fmt.Println("Processing complete!")
}

// getImageFiles returns a list of image files in the given directory
func getImageFiles(dir string) ([]string, error) {
	var files []string
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		
		// Check if it's an image file
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			files = append(files, filename)
		}
	}
	
	return files, nil
}

// copyFile copies a file from src to dest
func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	return nil
}
