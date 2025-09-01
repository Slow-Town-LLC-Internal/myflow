package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Environment represents the configuration for a cluster environment
type Environment struct {
	// AWS Configuration
	AWSProfile string `yaml:"aws_profile"`
	EKSCluster string `yaml:"eks_cluster"`
	EKSRegion  string `yaml:"eks_region"`
	
	// Authentication method: "aws-vault", "sso", or "none" (for local)
	AuthMethod string `yaml:"auth_method,omitempty"`
	
	// Local cluster configuration
	IsLocal        bool   `yaml:"is_local,omitempty"`
	LocalContext   string `yaml:"local_context,omitempty"`
	
	// Security settings
	RequireTouchID bool `yaml:"require_touchid,omitempty"`
	CacheDuration  string `yaml:"cache_duration,omitempty"` // e.g., "15m", "1h"
}

// Config represents the main configuration structure
type Config struct {
	Environments map[string]Environment `yaml:"environments"`
	
	// Global defaults
	DefaultAuthMethod     string `yaml:"default_auth_method,omitempty"`
	DefaultCacheDuration  string `yaml:"default_cache_duration,omitempty"`
}

// ClusterSwitcher handles cluster switching operations
type ClusterSwitcher struct {
	homeDir        string
	awsConfigPath  string
	kubeconfigPath string
	kubeconfigDir  string
	configPath     string
	config         *Config
}

// NewClusterSwitcher creates a new instance of ClusterSwitcher
func NewClusterSwitcher(configPath string) (*ClusterSwitcher, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	if configPath == "" {
		configPath = filepath.Join(homeDir, ".config", "eks-config.yaml")
	}

	cs := &ClusterSwitcher{
		homeDir:        homeDir,
		awsConfigPath:  filepath.Join(homeDir, ".aws", "config"),
		kubeconfigPath: filepath.Join(homeDir, ".kube", "config"),
		kubeconfigDir:  filepath.Join(homeDir, ".kube", "configs"),
		configPath:     configPath,
	}

	if err := cs.loadConfig(); err != nil {
		return nil, err
	}

	// Apply defaults
	if cs.config.DefaultAuthMethod == "" {
		cs.config.DefaultAuthMethod = "aws-vault"
	}
	if cs.config.DefaultCacheDuration == "" {
		cs.config.DefaultCacheDuration = "15m"
	}

	return cs, nil
}

// loadConfig loads and parses the configuration file
func (cs *ClusterSwitcher) loadConfig() error {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	cs.config = &Config{}
	if err := yaml.Unmarshal(data, cs.config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Apply defaults to each environment
	for name, env := range cs.config.Environments {
		if env.AuthMethod == "" && !env.IsLocal {
			env.AuthMethod = cs.config.DefaultAuthMethod
		}
		if env.CacheDuration == "" && !env.IsLocal {
			env.CacheDuration = cs.config.DefaultCacheDuration
		}
		// Local contexts don't use AWS auth
		if env.IsLocal {
			env.AuthMethod = "none"
		}
		cs.config.Environments[name] = env
	}

	return nil
}

// validateConfig ensures the configuration is valid for the given environment
func (cs *ClusterSwitcher) validateConfig(env string) error {
	if cs.config.Environments == nil {
		return fmt.Errorf("invalid config: 'environments' section missing")
	}

	envConfig, exists := cs.config.Environments[env]
	if !exists {
		return fmt.Errorf("environment '%s' not found in config", env)
	}

	// Local environments have different requirements
	if envConfig.IsLocal {
		if envConfig.LocalContext == "" {
			return fmt.Errorf("missing required config 'local_context' for local environment '%s'", env)
		}
		return nil
	}

	// AWS/EKS environments
	if envConfig.AWSProfile == "" {
		return fmt.Errorf("missing required config 'aws_profile' for environment '%s'", env)
	}

	if envConfig.EKSCluster == "" {
		return fmt.Errorf("missing required config 'eks_cluster' for environment '%s'", env)
	}

	return nil
}

// checkCommand checks if a command exists in PATH
func checkCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// runCommand executes a command and returns its output
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runCommandWithEnv executes a command with environment variables
func runCommandWithEnv(envVars map[string]string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	cmd.Env = os.Environ()
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	return cmd.Run()
}

// runCommandSilent executes a command silently and returns success/failure
func runCommandSilent(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	return cmd.Run() == nil
}

// verifyAWSVault checks if aws-vault is installed and configured
func (cs *ClusterSwitcher) verifyAWSVault(profile string) error {
	if !checkCommand("aws-vault") {
		return fmt.Errorf("aws-vault is not installed. Install it with: brew install aws-vault")
	}

	// Check if the profile exists in aws-vault
	cmd := exec.Command("aws-vault", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list aws-vault profiles: %v", err)
	}

	if !strings.Contains(string(output), profile) {
		fmt.Printf("Profile '%s' not found in aws-vault.\n", profile)
		fmt.Printf("To add credentials to aws-vault:\n")
		fmt.Printf("  aws-vault add %s\n", profile)
		fmt.Printf("\nThis will prompt for your AWS Access Key ID and Secret Access Key.\n")
		fmt.Printf("The credentials will be stored securely in macOS Keychain with TouchID protection.\n")
		return fmt.Errorf("profile '%s' not configured in aws-vault", profile)
	}

	return nil
}

// verifyAWSSession checks AWS credentials based on auth method
func (cs *ClusterSwitcher) verifyAWSSession(envConfig Environment) error {
	switch envConfig.AuthMethod {
	case "aws-vault":
		return cs.verifyAWSVault(envConfig.AWSProfile)
		
	case "sso":
		// Original SSO authentication
		cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", envConfig.AWSProfile)
		if err := cmd.Run(); err != nil {
			fmt.Printf("AWS SSO session expired. Initiating login...\n")
			if err := runCommand("aws", "sso", "login", "--profile", envConfig.AWSProfile); err != nil {
				return fmt.Errorf("failed to login to AWS SSO: %v", err)
			}
		}
		return nil
		
	case "none":
		// No authentication needed (local clusters)
		return nil
		
	default:
		return fmt.Errorf("unknown auth method: %s", envConfig.AuthMethod)
	}
}

// getKubeconfigPath returns the kubeconfig path for the given environment
func (cs *ClusterSwitcher) getKubeconfigPath(env string) string {
	envConfig := cs.config.Environments[env]
	if envConfig.IsLocal {
		return filepath.Join(cs.kubeconfigDir, fmt.Sprintf("local-%s", env), "config")
	}
	return filepath.Join(cs.kubeconfigDir, fmt.Sprintf("eks-%s", env), "config")
}

// ensureKubeconfigDir ensures the kubeconfig directory exists for the environment
func (cs *ClusterSwitcher) ensureKubeconfigDir(env string) error {
	envConfig := cs.config.Environments[env]
	var dirPath string
	
	if envConfig.IsLocal {
		dirPath = filepath.Join(cs.kubeconfigDir, fmt.Sprintf("local-%s", env))
	} else {
		dirPath = filepath.Join(cs.kubeconfigDir, fmt.Sprintf("eks-%s", env))
	}
	
	return os.MkdirAll(dirPath, 0755)
}

// updateEKSKubeconfig updates kubeconfig for EKS cluster
func (cs *ClusterSwitcher) updateEKSKubeconfig(env string, envConfig Environment) error {
	kubeconfigPath := cs.getKubeconfigPath(env)
	
	switch envConfig.AuthMethod {
	case "aws-vault":
		// Create a kubeconfig that uses aws-vault for authentication
		fmt.Printf("Updating kubeconfig to use aws-vault with TouchID...\n")
		
		// First, get the kubeconfig using aws-vault
		duration := envConfig.CacheDuration
		if duration == "" {
			duration = "15m"
		}
		
		// Generate the kubeconfig with aws-vault exec
		args := []string{
			"exec", envConfig.AWSProfile,
			"--duration=" + duration,
			"--",
			"aws", "eks", "update-kubeconfig",
			"--name", envConfig.EKSCluster,
			"--region", envConfig.EKSRegion,
			"--alias", fmt.Sprintf("eks-%s", env),
			"--kubeconfig", kubeconfigPath,
		}
		
		if err := runCommand("aws-vault", args...); err != nil {
			return fmt.Errorf("failed to update kubeconfig with aws-vault: %v", err)
		}
		
		// Now modify the kubeconfig to use aws-vault for future commands
		if err := cs.patchKubeconfigForAWSVault(kubeconfigPath, envConfig); err != nil {
			return fmt.Errorf("failed to patch kubeconfig for aws-vault: %v", err)
		}
		
	case "sso":
		// Original SSO-based update
		envVars := map[string]string{
			"KUBECONFIG": kubeconfigPath,
		}
		
		if err := runCommandWithEnv(envVars, "aws", "eks", "update-kubeconfig",
			"--name", envConfig.EKSCluster,
			"--profile", envConfig.AWSProfile,
			"--region", envConfig.EKSRegion,
			"--alias", fmt.Sprintf("eks-%s", env)); err != nil {
			return fmt.Errorf("failed to update kubeconfig: %v", err)
		}
		
	default:
		return fmt.Errorf("unsupported auth method for EKS: %s", envConfig.AuthMethod)
	}
	
	return nil
}

// patchKubeconfigForAWSVault modifies the kubeconfig to use aws-vault
func (cs *ClusterSwitcher) patchKubeconfigForAWSVault(kubeconfigPath string, envConfig Environment) error {
	// Read the current kubeconfig
	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %v", err)
	}
	
	var kubeconfig map[string]interface{}
	if err := yaml.Unmarshal(data, &kubeconfig); err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %v", err)
	}
	
	// Find and update the user exec command
	users, ok := kubeconfig["users"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid kubeconfig format: no users section")
	}
	
	for _, userInterface := range users {
		user, ok := userInterface.(map[interface{}]interface{})
		if !ok {
			continue
		}
		
		userData, ok := user["user"].(map[interface{}]interface{})
		if !ok {
			continue
		}
		
		exec, ok := userData["exec"].(map[interface{}]interface{})
		if !ok {
			continue
		}
		
		// Modify the exec command to use aws-vault
		exec["command"] = "aws-vault"
		
		duration := envConfig.CacheDuration
		if duration == "" {
			duration = "15m"
		}
		
		// Build new args for aws-vault
		args := []interface{}{
			"exec",
			envConfig.AWSProfile,
			"--duration=" + duration,
			"--",
			"aws",
		}
		
		// Preserve the original aws command args
		if originalArgs, ok := exec["args"].([]interface{}); ok {
			// Skip the profile args as aws-vault handles that
			for _, arg := range originalArgs {
				argStr := fmt.Sprintf("%v", arg)
				if !strings.HasPrefix(argStr, "--profile") {
					args = append(args, arg)
				}
			}
		}
		
		exec["args"] = args
		
		// Add environment variable to indicate TouchID requirement
		if envConfig.RequireTouchID {
			exec["env"] = []map[string]string{
				{"name": "AWS_VAULT_PROMPT", "value": "osascript"},
			}
		}
	}
	
	// Write the modified kubeconfig back
	modifiedData, err := yaml.Marshal(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to marshal modified kubeconfig: %v", err)
	}
	
	if err := os.WriteFile(kubeconfigPath, modifiedData, 0600); err != nil {
		return fmt.Errorf("failed to write modified kubeconfig: %v", err)
	}
	
	fmt.Printf("✅ Kubeconfig patched to use aws-vault with TouchID authentication\n")
	return nil
}

// switchToLocalCluster switches to a local cluster (e.g., Orbstack)
func (cs *ClusterSwitcher) switchToLocalCluster(env string, envConfig Environment) error {
	kubeconfigPath := cs.getKubeconfigPath(env)
	
	// Check if the local context exists
	cmd := exec.Command("kubectl", "config", "get-contexts", envConfig.LocalContext)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("local context '%s' not found", envConfig.LocalContext)
	}
	
	// Export the local context to a separate kubeconfig file
	envVars := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}
	
	// First, copy the current config
	defaultConfig := filepath.Join(cs.homeDir, ".kube", "config")
	configData, err := os.ReadFile(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to read default kubeconfig: %v", err)
	}
	
	if err := os.WriteFile(kubeconfigPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig: %v", err)
	}
	
	// Set the current context
	if err := runCommandWithEnv(envVars, "kubectl", "config", "use-context", envConfig.LocalContext); err != nil {
		return fmt.Errorf("failed to set context: %v", err)
	}
	
	return nil
}

// SwitchCluster switches to the specified environment cluster
func (cs *ClusterSwitcher) SwitchCluster(env string) error {
	if err := cs.validateConfig(env); err != nil {
		return err
	}

	envConfig := cs.config.Environments[env]

	// Ensure kubeconfig directory exists
	if err := cs.ensureKubeconfigDir(env); err != nil {
		return fmt.Errorf("failed to create kubeconfig directory: %v", err)
	}

	// Handle based on cluster type
	if envConfig.IsLocal {
		fmt.Printf("Switching to local cluster: %s\n", envConfig.LocalContext)
		if err := cs.switchToLocalCluster(env, envConfig); err != nil {
			return err
		}
	} else {
		// Verify AWS session/credentials
		if err := cs.verifyAWSSession(envConfig); err != nil {
			return err
		}

		// Update kubeconfig for EKS
		if err := cs.updateEKSKubeconfig(env, envConfig); err != nil {
			return err
		}

		// Set AWS_PROFILE for compatibility
		if err := os.Setenv("AWS_PROFILE", envConfig.AWSProfile); err != nil {
			return fmt.Errorf("failed to set AWS_PROFILE: %v", err)
		}
		
		fmt.Printf("AWS_PROFILE environment variable set to: %s\n", envConfig.AWSProfile)
	}

	// Set KUBECONFIG environment variable
	kubeconfigPath := cs.getKubeconfigPath(env)
	if err := os.Setenv("KUBECONFIG", kubeconfigPath); err != nil {
		return fmt.Errorf("failed to set KUBECONFIG: %v", err)
	}

	fmt.Printf("Successfully switched to %s environment\n", env)
	fmt.Printf("KUBECONFIG environment variable set to: %s\n", kubeconfigPath)

	// Show current context
	envVars := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}
	if err := runCommandWithEnv(envVars, "kubectl", "config", "current-context"); err != nil {
		return fmt.Errorf("failed to get current context: %v", err)
	}

	// Export commands for manual shell usage
	fmt.Printf("\nTo use this configuration in your current shell, run:\n")
	if !envConfig.IsLocal {
		fmt.Printf("export AWS_PROFILE=%s\n", envConfig.AWSProfile)
	}
	fmt.Printf("export KUBECONFIG=%s\n", kubeconfigPath)
	
	if envConfig.AuthMethod == "aws-vault" {
		fmt.Printf("\n💡 Tip: kubectl commands will now trigger TouchID authentication via aws-vault\n")
		fmt.Printf("   Credentials will be cached for %s\n", envConfig.CacheDuration)
	}

	return nil
}

// ListEnvironments lists all available environments
func (cs *ClusterSwitcher) ListEnvironments() {
	fmt.Println("Available environments:")
	for name, env := range cs.config.Environments {
		authInfo := ""
		if env.IsLocal {
			authInfo = " (local)"
		} else {
			authInfo = fmt.Sprintf(" (auth: %s)", env.AuthMethod)
		}
		fmt.Printf("  - %s%s\n", name, authInfo)
		if env.IsLocal {
			fmt.Printf("    Context: %s\n", env.LocalContext)
		} else {
			fmt.Printf("    Cluster: %s\n", env.EKSCluster)
			fmt.Printf("    Region: %s\n", env.EKSRegion)
			fmt.Printf("    Profile: %s\n", env.AWSProfile)
		}
		if env.RequireTouchID {
			fmt.Printf("    TouchID: required\n")
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cluster-switcher <environment|list>")
		fmt.Println("Examples:")
		fmt.Println("  cluster-switcher list    # List all environments")
		fmt.Println("  cluster-switcher prd     # Switch to production")
		fmt.Println("  cluster-switcher stg     # Switch to staging")
		fmt.Println("  cluster-switcher local   # Switch to local cluster")
		os.Exit(1)
	}

	configPath := os.Getenv("EKS_CONFIG")
	switcher, err := NewClusterSwitcher(configPath)
	if err != nil {
		fmt.Printf("Error initializing cluster switcher: %v\n", err)
		os.Exit(1)
	}

	if os.Args[1] == "list" {
		switcher.ListEnvironments()
		os.Exit(0)
	}

	if err := switcher.SwitchCluster(os.Args[1]); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}