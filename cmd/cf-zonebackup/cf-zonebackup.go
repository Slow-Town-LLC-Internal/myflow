package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"gopkg.in/yaml.v3"
)

// Config holds configuration loaded from YAML or environment
type Config struct {
	Email     string   `yaml:"email"`
	APIKey    string   `yaml:"api_key"`
	APIToken  string   `yaml:"api_token"`
	ZoneName  string   `yaml:"zone_name"`
	ZoneNames []string `yaml:"zone_names"`
}

// exportZone exports DNS records for a single zone
func exportZone(api *cloudflare.API, zoneName string) error {
	// Get the zone ID
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("error getting zone ID: %w", err)
	}

	// Get DNS records for the zone
	ctx := context.Background()
	records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return fmt.Errorf("error getting DNS records: %w", err)
	}

	// Create timestamp for filenames
	timestamp := time.Now().Format("20060102_150405")

	// Create output directory if it doesn't exist
	outputDir := "cloudflare-export"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Export to JSON
	jsonFilename := filepath.Join(outputDir, fmt.Sprintf("%s_%s.json", zoneName, timestamp))
	if err := exportToJSON(records, jsonFilename); err != nil {
		fmt.Printf("Error exporting to JSON: %v\n", err)
	} else {
		fmt.Printf("Exported %d records to %s\n", len(records), jsonFilename)
	}

	// Export to BIND zone format
	bindFilename := filepath.Join(outputDir, fmt.Sprintf("%s_%s.zone", zoneName, timestamp))
	if err := exportToBIND(records, bindFilename, zoneName); err != nil {
		fmt.Printf("Error exporting to BIND format: %v\n", err)
	} else {
		fmt.Printf("Exported %d records to BIND format file %s\n", len(records), bindFilename)
	}
	
	return nil
}

// exportAllZones exports all zones that the authenticated user has access to
func exportAllZones(api *cloudflare.API) (int, error) {
	ctx := context.Background()
	
	// Get all zones
	zones, err := api.ListZones(ctx)
	if err != nil {
		return 0, fmt.Errorf("error listing zones: %w", err)
	}
	
	if len(zones) == 0 {
		return 0, fmt.Errorf("no zones found for this account")
	}
	
	fmt.Printf("Found %d zones to export\n", len(zones))
	
	successCount := 0
	for _, zone := range zones {
		fmt.Printf("Exporting zone: %s\n", zone.Name)
		if err := exportZone(api, zone.Name); err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			successCount++
		}
	}
	
	return successCount, nil
}

// loadConfig loads configuration from YAML file and environment variables
func loadConfig() (Config, error) {
	config := Config{}

	// Try to load from config.yaml file
	configFile := findConfigFile()
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err == nil {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return config, fmt.Errorf("error parsing config file: %w", err)
			}
			fmt.Printf("Loaded configuration from %s\n", configFile)
		}
	}

	// Override with environment variables if present
	if os.Getenv("CLOUDFLARE_EMAIL") != "" {
		config.Email = os.Getenv("CLOUDFLARE_EMAIL")
	}
	if os.Getenv("CLOUDFLARE_API_KEY") != "" {
		config.APIKey = os.Getenv("CLOUDFLARE_API_KEY")
	}
	if os.Getenv("CLOUDFLARE_API_TOKEN") != "" {
		config.APIToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	}
	if os.Getenv("CLOUDFLARE_ZONE_NAME") != "" {
		config.ZoneName = os.Getenv("CLOUDFLARE_ZONE_NAME")
	}

	return config, nil
}

// findConfigFile looks for config.yaml in current and parent directories
func findConfigFile() string {
	candidates := []string{
		"config.yaml",
		"cloudflare.yaml",
		filepath.Join("config", "cloudflare.yaml"),
		filepath.Join("..", "config", "cloudflare.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "cloudflare.yaml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// exportToJSON exports DNS records to a JSON file
func exportToJSON(records []cloudflare.DNSRecord, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Using fmt.Fprintf for simplicity. In a production app, you'd want to use json.Marshal
	fmt.Fprintf(file, "[\n")
	for i, record := range records {
		fmt.Fprintf(file, "  {\n")
		fmt.Fprintf(file, "    \"type\": %q,\n", record.Type)
		fmt.Fprintf(file, "    \"name\": %q,\n", record.Name)
		fmt.Fprintf(file, "    \"content\": %q,\n", record.Content)
		fmt.Fprintf(file, "    \"ttl\": %d,\n", record.TTL)
		if record.Proxied != nil {
			fmt.Fprintf(file, "    \"proxied\": %t", *record.Proxied)
		} else {
			fmt.Fprintf(file, "    \"proxied\": false")
		}
		if record.Priority != nil {
			fmt.Fprintf(file, ",\n    \"priority\": %d", *record.Priority)
		}
		fmt.Fprintf(file, "\n  }")
		
		if i < len(records)-1 {
			fmt.Fprintf(file, ",")
		}
		fmt.Fprintf(file, "\n")
	}
	fmt.Fprintf(file, "]\n")

	return nil
}

// exportToBIND exports DNS records to a BIND zone file
func exportToBIND(records []cloudflare.DNSRecord, filename string, zoneName string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write zone file header
	fmt.Fprintf(file, "$ORIGIN %s.\n", zoneName)
	fmt.Fprintf(file, "$TTL 3600\n\n")

	for _, record := range records {
		// Skip Cloudflare-specific records that won't transfer to other providers
		if record.Proxied != nil && *record.Proxied && !isStandardRecord(record.Type) {
			continue
		}

		// Format name: strip zone suffix and use @ for root
		name := strings.TrimSuffix(record.Name, "."+zoneName)
		if name == zoneName || name == "" {
			name = "@"
		}

		// Use default TTL if set to Auto (0)
		ttl := record.TTL
		if ttl == 0 {
			ttl = 3600
		}

		// Format based on record type
		switch record.Type {
		case "MX", "SRV":
			priority := uint16(10)
			if record.Priority != nil {
				priority = *record.Priority
			}
			fmt.Fprintf(file, "%s %d IN %s %d %s\n", name, ttl, record.Type, priority, record.Content)
		default:
			fmt.Fprintf(file, "%s %d IN %s %s\n", name, ttl, record.Type, record.Content)
		}
	}

	return nil
}

// isStandardRecord returns true if the record type is a standard DNS record type
// that is likely to be supported by other DNS providers
func isStandardRecord(recordType string) bool {
	standardTypes := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CNAME": true,
		"TXT":   true,
		"MX":    true,
		"NS":    true,
		"SRV":   true,
		"CAA":   true,
		"PTR":   true,
	}
	return standardTypes[recordType]
}

func main() {
	// Initialize config from file and/or environment
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize Cloudflare API client
	var api *cloudflare.API
	if config.APIToken != "" {
		api, err = cloudflare.NewWithAPIToken(config.APIToken)
	} else if config.Email != "" && config.APIKey != "" {
		api, err = cloudflare.New(config.APIKey, config.Email)
	} else {
		fmt.Println("Error: Either API token or email+API key must be provided")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error initializing Cloudflare API: %v\n", err)
		os.Exit(1)
	}

	// Prepare zones to export
	var zonesToExport []string
	
	// Add single zone if specified
	if config.ZoneName != "" {
		zonesToExport = append(zonesToExport, config.ZoneName)
	}
	
	// Add zones from array if specified
	if len(config.ZoneNames) > 0 {
		zonesToExport = append(zonesToExport, config.ZoneNames...)
	}
	
	// Parse comma-separated CLOUDFLARE_ZONE_NAMES if present
	if zoneNamesEnv := os.Getenv("CLOUDFLARE_ZONE_NAMES"); zoneNamesEnv != "" {
		for _, zone := range strings.Split(zoneNamesEnv, ",") {
			zonesToExport = append(zonesToExport, strings.TrimSpace(zone))
		}
	}
	
	// Check if we have any zones to export
	if len(zonesToExport) == 0 {
		// If no zone specified, export all zones
		fmt.Println("No specific zones provided. Attempting to export all zones...")
		allZones, err := exportAllZones(api)
		if err != nil {
			fmt.Printf("Error exporting all zones: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully exported %d zones\n", allZones)
		return
	}
	
	// Export each specified zone
	for _, zoneName := range zonesToExport {
		fmt.Printf("Exporting zone: %s\n", zoneName)
		if err := exportZone(api, zoneName); err != nil {
			fmt.Printf("Error exporting zone %s: %v\n", zoneName, err)
		}
	}
}
