package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// --- Configuration ---

// stringSlice is a custom flag type for handling multiple string arguments
type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var (
	startPath    string
	excludePaths stringSlice
	includePaths stringSlice
)

func init() {
	// Define command-line flags
	flag.StringVar(&startPath, "start-path", "", "Directory to pack from, overriding Git repo detection.")
	flag.Var(&excludePaths, "exclude", "Pattern (glob) of files/directories to exclude. Can be used multiple times.")
	flag.Var(&excludePaths, "E", "Alias for -exclude.") // Alias
	flag.Var(&includePaths, "include", "Pattern (glob) of files/directories to include. If used, only matching files are packed. Can be used multiple times.")
	flag.Var(&includePaths, "I", "Alias for -include.") // Alias

	// Customize usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, `
Concatenates files into a single output, respecting Git structure or specified path.

Arguments:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Pack all tracked files in the current Git repo
  %s

  # Pack files from a specific directory, ignoring Git
  %s --start-path /path/to/project

  # Pack only Go files from the current Git repo
  %s --include "*.go"

  # Pack files, excluding logs and vendor directories (relative to root)
  %s --exclude "*.log" --exclude "vendor/*"

  # Pack specific Go files from a different project directory
  %s --start-path ../other-project --include "cmd/*.go" --include "pkg/utils/*.go"
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}
}

// --- Helper Functions ---

// isExcluded checks if a relative path matches any exclude patterns.
// Uses filepath.Match, expects patterns relative to rootPath.
func isExcluded(relativePath string) bool {
	// Ensure relativePath uses slashes for consistent matching
	matchPath := filepath.ToSlash(relativePath)
	for _, pattern := range excludePaths {
		// Ensure pattern uses slashes
		matchPattern := filepath.ToSlash(pattern)

		matched, err := filepath.Match(matchPattern, matchPath)
		if err != nil {
			log.Printf("Warning: Invalid exclude pattern '%s': %v", pattern, err)
			continue
		}
		if matched {
			// log.Printf("DEBUG: Path '%s' excluded by pattern '%s'", relativePath, pattern)
			return true
		}

		// Check if a directory pattern excludes this file
		// e.g., pattern "vendor/" should exclude "vendor/a.go"
		if strings.HasSuffix(matchPattern, "/") {
			if strings.HasPrefix(matchPath, matchPattern) {
				// log.Printf("DEBUG: Path '%s' excluded by directory pattern '%s'", relativePath, pattern)
				return true
			}
		}
		// e.g., pattern "build/*" should exclude "build/subdir/file.txt"
		if strings.HasSuffix(matchPattern, "/*") {
			dirPrefix := strings.TrimSuffix(matchPattern, "*") // Gives "build/"
			if strings.HasPrefix(matchPath, dirPrefix) {
				// log.Printf("DEBUG: Path '%s' excluded by wildcard directory pattern '%s'", relativePath, pattern)
				return true
			}
		}

		// Check parent directories (more robustly)
		currentDir := filepath.Dir(matchPath)
		for currentDir != "." && currentDir != "/" {
			// Check against patterns ending in / or /* specifically for directories
			for _, p := range excludePaths {
				pSlash := filepath.ToSlash(p)
				isDirPattern := strings.HasSuffix(pSlash, "/") || strings.HasSuffix(pSlash, "/*")
				if isDirPattern {
					patternToCheck := strings.TrimSuffix(strings.TrimSuffix(pSlash, "*"), "/")
					// Use filepath.Match on the directory itself
					dirMatched, _ := filepath.Match(patternToCheck, currentDir)
					if dirMatched {
						// log.Printf("DEBUG: Path '%s' excluded because parent dir '%s' matched pattern '%s'", relativePath, currentDir, p)
						return true
					}
				}
			}
			// Move up one directory, ensuring clean paths
			newDir := filepath.Dir(currentDir)
			if newDir == currentDir { // Avoid infinite loop at root
				break
			}
			currentDir = newDir
		}
	}
	return false
}

// isIncluded checks if a relative path matches any include patterns.
// If no include patterns are specified, all paths are considered included.
// Uses filepath.Match, expects patterns relative to rootPath.
func isIncluded(relativePath string) bool {
	if len(includePaths) == 0 {
		return true // Include everything if no specific includes are given
	}
	matchPath := filepath.ToSlash(relativePath)
	for _, pattern := range includePaths {
		matchPattern := filepath.ToSlash(pattern)
		matched, err := filepath.Match(matchPattern, matchPath)
		if err != nil {
			log.Printf("Warning: Invalid include pattern '%s': %v", pattern, err)
			continue
		}
		if matched {
			// log.Printf("DEBUG: Path '%s' included by pattern '%s'", relativePath, pattern)
			return true
		}
	}
	// log.Printf("DEBUG: Path '%s' not included by any pattern", relativePath)
	return false // Not included by any pattern
}

// processFile reads a file and writes its content with a header to the writer.
func processFile(writer io.Writer, fullPath, relativePath string) error {
	// 1. Final check: ensure it's not a symlink on the filesystem level
	//    (Should have been caught earlier, but safety first)
	fileInfo, err := os.Lstat(fullPath)
	if err != nil {
		log.Printf("Warning: Could not Lstat file '%s' just before processing: %v", fullPath, err)
		return nil // Skip if cannot stat
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		log.Printf("Skipping symlink confirmed by Lstat: %s", relativePath)
		return nil
	}
	if fileInfo.IsDir() {
		// Should not happen if called correctly, but handle defensively
		log.Printf("Skipping directory passed to processFile: %s", relativePath)
		return nil
	}

	// 2. Apply Filters based on RELATIVE path (using consistent slash separators)
	filterPath := filepath.ToSlash(relativePath)
	if !isIncluded(filterPath) {
		return nil
	}
	if isExcluded(filterPath) {
		return nil
	}

	// 3. Write Header
	// Use the original relativePath (might have OS-specific separators if from walk)
	// Standardize to slashes for header consistency.
	header := fmt.Sprintf("\n--- // File: %s // ---\n", filepath.ToSlash(relativePath))
	_, err = writer.Write([]byte(header))
	if err != nil {
		return fmt.Errorf("error writing header for %s: %w", relativePath, err)
	}

	// 4. Open and Write Content
	file, err := os.Open(fullPath)
	if err != nil {
		// Log warning but continue with other files
		log.Printf("Warning: Could not open file '%s': %v", fullPath, err)
		return nil // Skip this file
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		// Don't stop entirely, but log the error
		log.Printf("Warning: Error copying content for %s: %v", relativePath, err)
		return nil // Skip writing the trailing newline if copy failed
	}

	// Add a newline after file content for better separation
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		// This error is less critical but indicates potential output stream issues
		log.Printf("Warning: Error writing trailing newline for %s: %v", relativePath, err)
	}

	// fmt.Printf("DEBUG: Packed '%s'\n", relativePath)
	return nil // Indicate success for this file
}

// --- Main Logic ---

func main() {
	flag.Parse()

	var rootPath string
	var err error
	useGit := false

	// 1. Determine Root Path and Mode (Git vs Filesystem Walk)
	if startPath != "" {
		rootPath, err = filepath.Abs(startPath)
		if err != nil {
			log.Fatalf("Error getting absolute path for start-path '%s': %v", startPath, err)
		}
		info, err := os.Stat(rootPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("Error: start-path '%s' does not exist.", rootPath)
			}
			log.Fatalf("Error stating start-path '%s': %v", rootPath, err)
		}
		if !info.IsDir() {
			log.Fatalf("Error: start-path '%s' is not a directory.", rootPath)
		}
		log.Printf("Using specified start path: %s (Filesystem Walk Mode)", rootPath)

	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{DetectDotGit: true})
		if err == nil {
			worktree, err := repo.Worktree()
			if err != nil {
				log.Fatalf("Error getting worktree from Git repository: %v", err)
			}
			rootPath = worktree.Filesystem.Root()
			useGit = true
			log.Printf("Found Git repository root: %s (Git Mode)", rootPath)
		} else if err == git.ErrRepositoryNotExists {
			rootPath = cwd
			log.Printf("No Git repository found. Using current directory: %s (Filesystem Walk Mode)", rootPath)
		} else {
			log.Fatalf("Error opening Git repository: %v", err)
		}
	}

	rootPath = filepath.Clean(rootPath)
	writer := os.Stdout

	// 2. Process Files based on Mode
	if useGit {
		// --- Git Mode ---
		repo, err := git.PlainOpen(rootPath)
		if err != nil {
			log.Fatalf("Error re-opening git repository at '%s': %v", rootPath, err)
		}
		ref, err := repo.Head()
		if err != nil {
			// Handle case of repo with no commits yet
			if err.Error() == "reference not found" { // Or check specific error type if available
				log.Printf("Warning: Git repository has no commits (HEAD not found). No files to pack from Git index.")
				os.Exit(0) // Exit cleanly, nothing to do
			}
			log.Fatalf("Error getting HEAD reference: %v", err)
		}
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Fatalf("Error getting HEAD commit object: %v", err)
		}
		tree, err := commit.Tree()
		if err != nil {
			log.Fatalf("Error getting commit tree: %v", err)
		}

		fileIter := tree.Files()
		err = fileIter.ForEach(func(f *object.File) error {
			relativePath := filepath.Clean(f.Name) // Clean path from git
			fullPath := filepath.Join(rootPath, relativePath)

			// --- CORRECTED CHECKS ---
			// Skip Gitlinks (Submodules) based on Git mode
			const gitLinkMode = 0160000 // Octal representation for GitLink
			const symlinkMode = 0120000 // Octal representation for Symlink

			// Skip Gitlinks (Submodules) based on Git mode using octal value
			if f.Mode == filemode.FileMode(gitLinkMode) { // Cast octal to filemode.FileMode
				log.Printf("Skipping submodule (gitlink): %s", relativePath)
				return nil
			}
			// Skip Symlinks based on Git mode using octal value
			if f.Mode == filemode.FileMode(symlinkMode) { // Cast octal to filemode.FileMode
				log.Printf("Skipping symlink (git mode): %s", relativePath)
				return nil
			}

			// --- End Corrected Checks ---

			// Additionally check the file system state for safety
			// os.Lstat does not follow symlinks
			fileInfo, statErr := os.Lstat(fullPath)
			if statErr != nil {
				// File might be in the index but modified/deleted in worktree
				log.Printf("Skipping file not found or inaccessible in worktree: %s (%v)", relativePath, statErr)
				return nil // Skip this file
			}

			// Check if it's a symlink on the filesystem, even if git thinks it's regular
			// (Could happen with core.symlinks=false config or manual changes)
			if fileInfo.Mode()&os.ModeSymlink != 0 {
				log.Printf("Skipping symlink (fs check): %s", relativePath)
				return nil
			}

			// Ensure it's not actually a directory in the worktree
			if fileInfo.IsDir() {
				// This usually indicates something inconsistent, like a file replaced by a dir
				log.Printf("Skipping directory found in worktree at expected file path: %s", relativePath)
				return nil
			}

			// If we got here, it should be a file-like object we want.
			// processFile will do the include/exclude filtering based on relativePath
			// Use filepath.ToSlash for consistent path separators in filters and headers
			return processFile(writer, fullPath, filepath.ToSlash(relativePath))
		})
		if err != nil {
			log.Fatalf("Error iterating through Git files: %v", err)
		}

	} else {
		// --- Filesystem Walk Mode ---
		err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				log.Printf("Warning: Error accessing path %q during walk: %v", path, walkErr)
				// Decide if the error is fatal for the walk or just for this path
				if d != nil && d.IsDir() {
					return filepath.SkipDir // Skip directory if cannot access
				}
				return nil // Skip file entry if error
			}

			relativePath, err := filepath.Rel(rootPath, path)
			if err != nil {
				log.Printf("Warning: Could not get relative path for %q: %v", path, err)
				return nil // Skip if relative path calculation fails
			}

			// Use slashes for internal consistency (matching, filtering)
			relativePathSlash := filepath.ToSlash(relativePath)

			if relativePathSlash == "." {
				return nil // Skip root directory entry
			}

			// Skip symlinks
			if d.Type()&fs.ModeSymlink != 0 {
				log.Printf("Skipping symlink: %s", relativePathSlash)
				return nil
			}

			// Handle directories: check exclusion and skip if needed
			if d.IsDir() {
				// Check if the directory itself or a pattern like "dir/*" excludes it
				if isExcluded(relativePathSlash+"/") || isExcluded(relativePathSlash) {
					log.Printf("Skipping excluded directory: %s", relativePathSlash)
					return filepath.SkipDir // Don't descend into excluded directories
				}
				return nil // It's a directory, not excluded, continue walking
			}

			// Process regular files (processFile handles include/exclude)
			// Pass the original 'relativePath' in case the user expects OS-specific separators somewhere downstream,
			// although processFile standardizes to slashes for header/filtering.
			return processFile(writer, path, relativePath)
		})
		if err != nil {
			log.Fatalf("Error walking directory '%s': %v", rootPath, err)
		}
	}

	log.Println("File packing completed.")
}
