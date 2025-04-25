package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/google/uuid"
)

const (
	defaultRegion = "us-central-2" // Or your most common region
	opCreate      = "create"
	opRotate      = "rotate" // Overwrite existing
	opDelete      = "delete"
	opGet         = "get"    // Added for verification
	keyTypeES256  = "ES256"
	// Add other key types here later, e.g., keyTypeRSA256, keyTypePEM
)

// Config holds command line arguments
type Config struct {
	Operation   string
	Env         string
	PathSuffix  string // e.g., /papi/auth-jwk-set
	KeyType     string
	Value       string // Pre-defined value (optional)
	Profile     string // Specific AWS profile (optional)
	Region      string
	KmsKeyID    string // KMS Key for SecureString (optional)
	SkipConfirm bool   // Skip confirmation prompts
}

func main() {
	cfg := parseFlags()
	ctx := context.Background()

	awsCfg, err := loadAWSConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)
	fullPath := fmt.Sprintf("/myorg/%s%s", cfg.Env, cfg.PathSuffix)

	fmt.Printf("Operation:    %s\n", cfg.Operation)
	fmt.Printf("Environment:  %s\n", cfg.Env)
	fmt.Printf("AWS Profile:  %s\n", getProfileName(cfg))
	fmt.Printf("AWS Region:   %s\n", awsCfg.Region)
	fmt.Printf("SSM Path:     %s\n", fullPath)
	if cfg.KmsKeyID != "" {
		fmt.Printf("KMS Key ID:   %s\n", cfg.KmsKeyID)
	}
	fmt.Println("---")

	switch cfg.Operation {
	case opCreate:
		handleCreate(ctx, ssmClient, cfg, fullPath)
	case opRotate:
		handleRotate(ctx, ssmClient, cfg, fullPath)
	case opDelete:
		handleDelete(ctx, ssmClient, cfg, fullPath)
	case opGet:
		handleGet(ctx, ssmClient, cfg, fullPath)
	default:
		log.Fatalf("Invalid operation: %s. Must be one of: %s, %s, %s, %s",
			cfg.Operation, opCreate, opRotate, opDelete, opGet)
	}
}

// --- Flag Parsing ---

func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.Operation, "operation", "", fmt.Sprintf("Operation to perform (%s, %s, %s, %s)", opCreate, opRotate, opDelete, opGet))
	flag.StringVar(&cfg.Env, "env", "", "Target environment (e.g., stg, prd)")
	flag.StringVar(&cfg.PathSuffix, "path", "", "SSM parameter path suffix (e.g., /redis/auth-password or /papi/auth-jwk-set)")
	flag.StringVar(&cfg.KeyType, "key-type", "", fmt.Sprintf("Type of key to generate if --value is not provided (e.g., %s)", keyTypeES256))
	flag.StringVar(&cfg.Value, "value", "", "Provide a specific value instead of generating a key (use for non-key secrets)")
	flag.StringVar(&cfg.Profile, "profile", "", "Specific AWS profile to use (overrides default/env derivation)")
	flag.StringVar(&cfg.Region, "region", defaultRegion, "AWS Region")
	flag.StringVar(&cfg.KmsKeyID, "kms-key-id", "", "KMS Key ID or ARN for SecureString encryption (optional)")
	flag.BoolVar(&cfg.SkipConfirm, "yes", false, "Skip confirmation prompts for rotate/delete operations")
	flag.Parse()

	// Validation
	if cfg.Operation == "" || cfg.Env == "" || cfg.PathSuffix == "" {
		flag.Usage()
		log.Fatal("Error: --operation, --env, and --path are required.")
	}
	if (cfg.Operation == opCreate || cfg.Operation == opRotate) && cfg.Value == "" && cfg.KeyType == "" {
		flag.Usage()
		log.Fatal("Error: --key-type must be specified for create/rotate operations if --value is not provided.")
	}
	if cfg.Value != "" && cfg.KeyType != "" {
		log.Println("Warning: Both --value and --key-type provided. --value will be used.")
		cfg.KeyType = "" // Ignore key-type if value is explicitly given
	}

	return cfg
}

// --- AWS Configuration ---

func loadAWSConfig(ctx context.Context, cfg Config) (aws.Config, error) {
	profileName := getProfileName(cfg)
	options := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithSharedConfigProfile(profileName),
	}
	return config.LoadDefaultConfig(ctx, options...)
}

func getProfileName(cfg Config) string {
	if cfg.Profile != "" {
		return cfg.Profile
	}
	// Default profile naming convention
	return fmt.Sprintf("myorg-%s-admin", cfg.Env)
}

// --- Operation Handlers ---

func handleCreate(ctx context.Context, client *ssm.Client, cfg Config, fullPath string) {
	_, err := getSSMParameter(ctx, client, fullPath)
	if err == nil {
		log.Fatalf("Error: Parameter %s already exists. Use 'rotate' operation to overwrite.", fullPath)
	}
	var paramNotFound *ssmtypes.ParameterNotFound
	if !errors.As(err, &paramNotFound) {
		log.Fatalf("Error checking for existing parameter %s: %v", fullPath, err)
	}

	// Parameter doesn't exist, proceed with creation
	value, details := getValueForCreateOrRotate(cfg)
	err = putSSMParameter(ctx, client, fullPath, value, cfg.KmsKeyID, false) // Overwrite=false
	if err != nil {
		log.Fatalf("Failed to create parameter %s: %v", fullPath, err)
	}
	fmt.Printf("Successfully created parameter: %s\n", fullPath)
	if details != "" {
		fmt.Printf("Generated Key Details: %s\n", details)
	}
}

func handleRotate(ctx context.Context, client *ssm.Client, cfg Config, fullPath string) {
	// Check if parameter exists (optional, but good practice)
	_, err := getSSMParameter(ctx, client, fullPath)
	if err != nil {
		var paramNotFound *ssmtypes.ParameterNotFound
		if errors.As(err, &paramNotFound) {
			log.Printf("Warning: Parameter %s does not exist, but creating it with 'rotate' operation.", fullPath)
			// Allow creation even if rotate was specified but param doesn't exist
		} else {
			log.Fatalf("Error checking for existing parameter %s: %v", fullPath, err)
		}
	}

	if !cfg.SkipConfirm {
		if !promptConfirm(fmt.Sprintf("Overwrite parameter %s?", fullPath)) {
			log.Println("Operation cancelled by user.")
			return
		}
	}

	value, details := getValueForCreateOrRotate(cfg)
	err = putSSMParameter(ctx, client, fullPath, value, cfg.KmsKeyID, true) // Overwrite=true
	if err != nil {
		log.Fatalf("Failed to rotate/overwrite parameter %s: %v", fullPath, err)
	}
	fmt.Printf("Successfully rotated/overwritten parameter: %s\n", fullPath)
	if details != "" {
		fmt.Printf("Generated Key Details: %s\n", details)
	}
}

func handleDelete(ctx context.Context, client *ssm.Client, cfg Config, fullPath string) {
	// Verify parameter exists before attempting delete
	_, err := getSSMParameter(ctx, client, fullPath)
	if err != nil {
		var paramNotFound *ssmtypes.ParameterNotFound
		if errors.As(err, &paramNotFound) {
			log.Fatalf("Error: Parameter %s does not exist. Cannot delete.", fullPath)
		}
		log.Fatalf("Error checking for parameter %s before delete: %v", fullPath, err)
	}

	if !cfg.SkipConfirm {
		if !promptConfirm(fmt.Sprintf("Permanently delete parameter %s?", fullPath)) {
			log.Println("Operation cancelled by user.")
			return
		}
	}

	err = deleteSSMParameter(ctx, client, fullPath)
	if err != nil {
		log.Fatalf("Failed to delete parameter %s: %v", fullPath, err)
	}
	fmt.Printf("Successfully deleted parameter: %s\n", fullPath)
}

func handleGet(ctx context.Context, client *ssm.Client, cfg Config, fullPath string) {
	paramOutput, err := getSSMParameter(ctx, client, fullPath)
	if err != nil {
		var paramNotFound *ssmtypes.ParameterNotFound
		if errors.As(err, &paramNotFound) {
			log.Fatalf("Error: Parameter %s not found.", fullPath)
		}
		log.Fatalf("Error getting parameter %s: %v", fullPath, err)
	}

	fmt.Println("Parameter found:")
	fmt.Printf("  Name: %s\n", aws.ToString(paramOutput.Parameter.Name))
	fmt.Printf("  Type: %s\n", paramOutput.Parameter.Type)
	fmt.Printf("  Version: %d\n", paramOutput.Parameter.Version)
	fmt.Printf("  Last Modified: %s\n", paramOutput.Parameter.LastModifiedDate.Format("2025-01-02 15:04:05 MST"))
	// Optionally print value - be careful with sensitive data
	// fmt.Printf("  Value: %s\n", aws.ToString(paramOutput.Parameter.Value))
	fmt.Println("  Value: <retrieved successfully - not displayed>")
}


// --- Value Generation ---

func getValueForCreateOrRotate(cfg Config) (value string, details string) {
	if cfg.Value != "" {
		return cfg.Value, "(User-provided value)"
	}

	switch cfg.KeyType {
	case keyTypeES256:
		jwkSetBytes, kid, err := generateES256JWKSet()
		if err != nil {
			log.Fatalf("Failed to generate %s key: %v", keyTypeES256, err)
		}
		return string(jwkSetBytes), fmt.Sprintf("(Generated %s JWK Set, kid: %s)", keyTypeES256, kid)
	// Add cases for other key types here
	// case keyTypeRSA256: ...
	// case keyTypePEM: ...
	default:
		log.Fatalf("Unsupported --key-type for generation: %s", cfg.KeyType)
		return "", "" // Should not be reached
	}
}

// --- Key Generation (JWK ES256 Example) ---

func generateES256JWKSet() (jwkSetJSON []byte, kid string, err error) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate EC key: %w", err)
	}

	publicKey := privateKey.PublicKey
	x := publicKey.X.Bytes()
	y := publicKey.Y.Bytes()
	d := privateKey.D.Bytes()

	// Ensure components are 32 bytes for P-256, padding if needed (rare for stdlib)
	x = padBytes(x, 32)
	y = padBytes(y, 32)
	d = padBytes(d, 32)


	keyID := uuid.NewString()

	jwk := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(x),
		"y":   base64.RawURLEncoding.EncodeToString(y),
		"d":   base64.RawURLEncoding.EncodeToString(d), // Private component
		"use": "sig",
		"alg": "ES256",
		"kid": keyID,
	}

	jwkSet := map[string]interface{}{
		"keys": []interface{}{jwk},
	}

	jwkSetBytes, err := json.MarshalIndent(jwkSet, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JWK Set to JSON: %w", err)
	}

	return jwkSetBytes, keyID, nil
}

// padBytes ensures byte slice has target length by padding with leading zeros
func padBytes(b []byte, length int) []byte {
    if len(b) >= length {
        // If longer (unlikely for stdlib keys), truncate or handle appropriately
        // For now, just return if equal or longer
        return b[len(b)-length:] // Return last 'length' bytes if too long
    }
    padded := make([]byte, length)
    copy(padded[length-len(b):], b)
    return padded
}


// --- SSM Interactions ---

func getSSMParameter(ctx context.Context, client *ssm.Client, name string) (*ssm.GetParameterOutput, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	}
	return client.GetParameter(ctx, input)
}

func putSSMParameter(ctx context.Context, client *ssm.Client, name, value, kmsKeyID string, overwrite bool) error {
	paramType := ssmtypes.ParameterTypeSecureString
	input := &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Type:      paramType,
		Overwrite: aws.Bool(overwrite),
	}
	if kmsKeyID != "" {
		input.KeyId = aws.String(kmsKeyID)
	}

	_, err := client.PutParameter(ctx, input)
	return err // Returns nil on success
}

func deleteSSMParameter(ctx context.Context, client *ssm.Client, name string) error {
	input := &ssm.DeleteParameterInput{
		Name: aws.String(name),
	}
	_, err := client.DeleteParameter(ctx, input)
	return err
}

// --- User Interaction ---

func promptConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [yes/no]: ", prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "yes" {
			return true
		}
		if input == "no" {
			return false
		}
	}
}
