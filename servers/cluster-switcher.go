package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Environment represents the configuration for an EKS environment
type Environment struct {
	AWSProfile string `yaml:"aws_profile"`
	EKSCluster string `yaml:"eks_cluster"`
	EKSRegion  string `yaml:"eks_region"`
}

// Config represents the main configuration structure
type Config struct {
	Environments map[string]Environment `yaml:"environments"`
}

// ClusterSwitcher handles EKS cluster switching operations
type ClusterSwitcher struct {
	homeDir        string
	awsConfigPath  string
	kubeconfigPath string
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
		configPath:     configPath,
	}

	if err := cs.loadConfig(); err != nil {
		return nil, err
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

	if envConfig.AWSProfile == "" {
		return fmt.Errorf("missing required config 'aws_profile' for environment '%s'", env)
	}

	if envConfig.EKSCluster == "" {
		return fmt.Errorf("missing required config 'eks_cluster' for environment '%s'", env)
	}

	return nil
}

// runCommand executes a command and returns its output
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// verifyAWSSession checks if the AWS SSO session is valid
func (cs *ClusterSwitcher) verifyAWSSession(profile string) error {
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profile)
	if err := cmd.Run(); err != nil {
		fmt.Printf("AWS SSO session expired. Initiating login...\n")
		if err := runCommand("aws", "sso", "login", "--profile", profile); err != nil {
			return fmt.Errorf("failed to login to AWS SSO: %v", err)
		}
	}
	return nil
}

// SwitchCluster switches to the specified environment cluster
func (cs *ClusterSwitcher) SwitchCluster(env string) error {
	if env != "prd" && env != "stg" {
		return fmt.Errorf("environment must be 'prd' or 'stg'")
	}

	if err := cs.validateConfig(env); err != nil {
		return err
	}

	envConfig := cs.config.Environments[env]

	// Verify AWS SSO session
	if err := cs.verifyAWSSession(envConfig.AWSProfile); err != nil {
		return err
	}

	// Update kubeconfig
	if err := runCommand("aws", "eks", "update-kubeconfig",
		"--name", envConfig.EKSCluster,
		"--profile", envConfig.AWSProfile,
		"--region", envConfig.EKSRegion); err != nil {
		return fmt.Errorf("failed to update kubeconfig: %v", err)
	}

	// Set AWS_PROFILE environment variable
	if err := os.Setenv("AWS_PROFILE", envConfig.AWSProfile); err != nil {
		return fmt.Errorf("failed to set AWS_PROFILE: %v", err)
	}

	fmt.Printf("Successfully switched to %s cluster: %s\n", env, envConfig.EKSCluster)
	fmt.Printf("AWS_PROFILE environment variable set to: %s\n", envConfig.AWSProfile)

	// Show current context
	if err := runCommand("kubectl", "config", "current-context"); err != nil {
		return fmt.Errorf("failed to get current context: %v", err)
	}

	return nil
}

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "prd" && os.Args[1] != "stg") {
		fmt.Println("Usage: cluster-switcher <prd|stg>")
		os.Exit(1)
	}

	configPath := os.Getenv("EKS_CONFIG")
	switcher, err := NewClusterSwitcher(configPath)
	if err != nil {
		fmt.Printf("Error initializing cluster switcher: %v\n", err)
		os.Exit(1)
	}

	if err := switcher.SwitchCluster(os.Args[1]); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
