package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Repository represents a GitHub repository
type Repository struct {
	Name          string `json:"name"`
	SSHURL        string `json:"ssh_url"`
	DefaultBranch string `json:"default_branch"`
}

func main() {
	// Command line flags
	orgName := flag.String("org", "", "GitHub organization name")
	userName := flag.String("user", "", "GitHub username for personal repositories")
	outputDir := flag.String("output", "", "Base output directory for cloned repositories (default: current directory)")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent clones")
	token := flag.String("token", "", "GitHub personal access token (optional, increases rate limits)")
	flag.Parse()

	// Validate input - either org or user must be specified
	if *orgName == "" && *userName == "" {
		fmt.Println("Error: Either GitHub organization name (-org) or username (-user) is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set default output directory if not specified
	if *outputDir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current directory: %v", err)
		}
		*outputDir = currentDir
	}

	var repos []Repository
	var err error
	var targetName string

	// Determine whether to fetch org repos or user repos
	if *orgName != "" {
		targetName = *orgName
		repos, err = getOrgRepositories(targetName, *token)
		if err != nil {
			log.Fatalf("Error fetching organization repositories: %v", err)
		}
		fmt.Printf("Found %d repositories in the %s organization\n", len(repos), targetName)
	} else {
		targetName = *userName
		repos, err = getUserRepositories(targetName, *token)
		if err != nil {
			log.Fatalf("Error fetching user repositories: %v", err)
		}
		fmt.Printf("Found %d repositories for user %s\n", len(repos), targetName)
	}

	// Create directory structure
	repoDir := filepath.Join(*outputDir, targetName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		log.Fatalf("Error creating directory for %s: %v", targetName, err)
	}

	// Clone repositories with concurrency limit
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *concurrency)

	for _, repo := range repos {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire

		go func(repo Repository) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release

			repoPath := filepath.Join(repoDir, repo.Name)
			cloneRepo(repo, repoPath)
		}(repo)
	}

	wg.Wait()
	fmt.Println("All repositories have been cloned successfully!")
}

// getOrgRepositories fetches all repositories for the specified organization
func getOrgRepositories(orgName, token string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?page=%d&per_page=%d", orgName, page, perPage)
		
		repos, hasMore, err := fetchRepos(url, token)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)
		
		if !hasMore {
			break
		}
		
		page++
	}

	return allRepos, nil
}

// getUserRepositories fetches all repositories for the specified user
func getUserRepositories(userName, token string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/users/%s/repos?page=%d&per_page=%d", userName, page, perPage)
		
		repos, hasMore, err := fetchRepos(url, token)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)
		
		if !hasMore {
			break
		}
		
		page++
	}

	return allRepos, nil
}

// fetchRepos makes the API request and returns the parsed repositories
func fetchRepos(url, token string) ([]Repository, bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("API request failed with status code %d: %s\nURL: %s", 
			resp.StatusCode, string(body), url)
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, false, err
	}

	// Check if there might be more pages
	hasMore := len(repos) > 0 && len(repos) == 100
	
	// Handle rate limiting
	remainingRequests := resp.Header.Get("X-RateLimit-Remaining")
	if remainingRequests == "0" {
		resetTime := resp.Header.Get("X-RateLimit-Reset")
		resetTimeInt, err := strconv.ParseInt(resetTime, 10, 64)
		if err == nil {
			resetTimeUnix := time.Unix(resetTimeInt, 0)
			waitTime := time.Until(resetTimeUnix)
			if waitTime > 0 {
				fmt.Printf("Rate limit reached. Waiting for %v seconds...\n", waitTime.Seconds())
				time.Sleep(waitTime)
			}
		}
	}

	return repos, hasMore, nil
}

// cloneRepo clones a repository to the specified directory
func cloneRepo(repo Repository, destDir string) {
	fmt.Printf("Cloning %s...\n", repo.Name)

	// Check if directory already exists
	if _, err := os.Stat(destDir); err == nil {
		fmt.Printf("Directory %s already exists, skipping clone of %s\n", destDir, repo.Name)
		return
	}

	// Create the git clone command
	cmd := exec.Command("git", "clone", repo.SSHURL, destDir)
	
	// Redirect stdout and stderr to capture output
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		fmt.Printf("Error cloning %s: %v\n%s\n", repo.Name, err, strings.TrimSpace(string(output)))
		return
	}
	
	fmt.Printf("Successfully cloned %s to %s\n", repo.Name, destDir)
}
