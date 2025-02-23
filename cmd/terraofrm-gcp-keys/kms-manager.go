package main

import (
  "bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	storage "cloud.google.com/go/storage"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type Config struct {
	projectID    string
	location     string
	keyRing      string
	keyName      string
	keyID        string
	envFile      string
	bucketName   string
	secretName   string
}

type TerraformOutput struct {
	KeyID               string
	KeyName             string
	KeyringName         string
	ServiceAccountEmail string
}

func parseTerraformOutput(filename string) (*TerraformOutput, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open terraform output file: %v", err)
	}
	defer file.Close()

	output := &TerraformOutput{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		// Remove quotes and leading/trailing spaces
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

		switch key {
		case "key_id":
			output.KeyID = value
		case "key_name":
			output.KeyName = value
		case "keyring_name":
			output.KeyringName = value
		case "service_account_email":
			output.ServiceAccountEmail = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading terraform output: %v", err)
	}

	// Validate required fields
	if output.KeyID == "" {
		return nil, fmt.Errorf("key_id not found in terraform output")
	}
	if output.KeyName == "" {
		return nil, fmt.Errorf("key_name not found in terraform output")
	}
	if output.KeyringName == "" {
		return nil, fmt.Errorf("keyring_name not found in terraform output")
	}

	return output, nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func main() {
	action := flag.String("action", "", "Action to perform: encrypt or decrypt")
	envFile := flag.String("env", "~/.env", "Environment file path")
	secretName := flag.String("secret", "env_secrets", "Name of the secret in GCS")
	tfOutput := flag.String("tf-output", "tf-output.log", "Terraform output file")
	flag.Parse()

	if *action == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Read and parse terraform output
	tfConfig, err := parseTerraformOutput(*tfOutput)
	if err != nil {
		log.Fatalf("Failed to parse terraform output: %v", err)
	}

	projectID, location := parseKeyIDComponents(tfConfig.KeyID)
	if projectID == "" || location == "" {
		log.Fatalf("Failed to parse project ID or location from key ID")
	}

	cfg := &Config{
		projectID:  projectID,
		location:   location,
		keyRing:    tfConfig.KeyringName,
		keyName:    tfConfig.KeyName,
		keyID:      tfConfig.KeyID,
		envFile:    *envFile,
		bucketName: fmt.Sprintf("%s-encrypted-secrets", projectID),
		secretName: *secretName,
	}

	ctx := context.Background()

	// Create KMS client
	kmsClient, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create KMS client: %v", err)
	}
	defer kmsClient.Close()

	// Create Storage client
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Storage client: %v", err)
	}
	defer storageClient.Close()

	switch *action {
	case "encrypt":
		if err := encryptAndUpload(ctx, kmsClient, storageClient, cfg); err != nil {
			log.Fatalf("Encryption failed: %v", err)
		}
	case "decrypt":
		if err := downloadAndDecrypt(ctx, kmsClient, storageClient, cfg); err != nil {
			log.Fatalf("Decryption failed: %v", err)
		}
	default:
		log.Fatalf("Invalid action: %s", *action)
	}
}

func encryptAndUpload(ctx context.Context, kmsClient *kms.KeyManagementClient, storageClient *storage.Client, cfg *Config) error {
	// Expand file path
	expandedPath, err := expandPath(cfg.envFile)
	if err != nil {
		return fmt.Errorf("failed to expand file path: %v", err)
	}

	// Read the environment file
	plaintext, err := ioutil.ReadFile(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read env file: %v", err)
	}

	// Encrypt the data
	req := &kmspb.EncryptRequest{
		Name:      cfg.keyID,
		Plaintext: plaintext,
	}

	result, err := kmsClient.Encrypt(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %v", err)
	}

	// Upload to GCS
	bucket := storageClient.Bucket(cfg.bucketName)
	obj := bucket.Object(cfg.secretName)

	writer := obj.NewWriter(ctx)
	if _, err := writer.Write(result.Ciphertext); err != nil {
		return fmt.Errorf("failed to write to GCS: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %v", err)
	}

	log.Printf("Successfully encrypted and uploaded %s to gs://%s/%s", cfg.envFile, cfg.bucketName, cfg.secretName)
	return nil
}

func createBackup(envPath string) error {
	// Check if file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil // No backup needed
	}

	// Create backup with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.%s.bak", envPath, timestamp)

	input, err := ioutil.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read existing env file: %v", err)
	}

	if err := ioutil.WriteFile(backupPath, input, 0600); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	log.Printf("Created backup at %s", backupPath)
	return nil
}

func downloadAndDecrypt(ctx context.Context, kmsClient *kms.KeyManagementClient, storageClient *storage.Client, cfg *Config) error {
	// Download from GCS
	bucket := storageClient.Bucket(cfg.bucketName)
	obj := bucket.Object(cfg.secretName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GCS reader: %v", err)
	}
	defer reader.Close()

	ciphertext, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read from GCS: %v", err)
	}

	// Decrypt the data
	req := &kmspb.DecryptRequest{
		Name:       cfg.keyID,
		Ciphertext: ciphertext,
	}

	result, err := kmsClient.Decrypt(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %v", err)
	}

	// Expand and write to env file
	expandedPath, err := expandPath(cfg.envFile)
	if err != nil {
		return fmt.Errorf("failed to expand file path: %v", err)
	}

	// Create backup if file exists
	if err := createBackup(expandedPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := ioutil.WriteFile(expandedPath, result.Plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write env file: %v", err)
	}

	log.Printf("Successfully downloaded and decrypted secrets to %s", expandedPath)
	return nil
}

func parseKeyIDComponents(keyID string) (string, string) {
	// Clean up the input string
	keyID = strings.Trim(keyID, "\"")
	parts := strings.Split(keyID, "/")

	// Extract project ID and location
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			projectID := parts[i+1]
			if i+3 < len(parts) && parts[i+2] == "locations" {
				location := parts[i+3]
				return projectID, location
			}
		}
	}
	return "", ""
}
