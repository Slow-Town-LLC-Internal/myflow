package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Command line arguments
var (
	awsProfile   string
	awsRegion    string
	logGroupStr  string
	monthsBack   int
	sortBySize   bool
	minSizeMB    float64
	detailedView bool
	outputCSV    string
)

func init() {
	flag.StringVar(&awsProfile, "aws-profile", "", "AWS profile to use")
	flag.StringVar(&awsRegion, "aws-region", "us-east-1", "AWS region")
	flag.StringVar(&logGroupStr, "log-groups", "", "Comma-separated list of app log groups to track separately")
	flag.IntVar(&monthsBack, "months-back", 1, "Number of months back to calculate (default: 1)")
	flag.BoolVar(&sortBySize, "sort", false, "Sort log groups by size (largest first)")
	flag.Float64Var(&minSizeMB, "min-size", 0, "Only show log groups larger than this size (in MB)")
	flag.BoolVar(&detailedView, "detailed", false, "Show detailed metrics including storage costs")
	flag.StringVar(&outputCSV, "csv", "", "Output results to CSV file (optional)")
	flag.Parse()
}

// LogGroupMetrics holds metrics for a log group
type LogGroupMetrics struct {
	Name          string
	IngestedBytes float64
	StoredBytes   float64
	IsAppLog      bool
}

func main() {
	// Create AWS session
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	if awsProfile != "" {
		opts.Profile = awsProfile
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		fmt.Printf("Error creating AWS session: %v\n", err)
		os.Exit(1)
	}

	cwLogs := cloudwatchlogs.New(sess, aws.NewConfig().WithRegion(awsRegion))
	cw := cloudwatch.New(sess, aws.NewConfig().WithRegion(awsRegion))

	// Calculate time range for last month
	now := time.Now()

	// Calculate first day of the target month
	firstDay := now.AddDate(0, -monthsBack, 0)
	firstDay = time.Date(firstDay.Year(), firstDay.Month(), 1, 0, 0, 0, 0, now.Location())

	// Calculate last day of the target month
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second)

	fmt.Printf("Calculating log sizes from %s to %s\n",
		firstDay.Format("2006-01-02"),
		lastDay.Format("2006-01-02"))

	// Get app log groups from parameter
	appLogGroups := make(map[string]bool)
	if logGroupStr != "" {
		for _, group := range strings.Split(logGroupStr, ",") {
			appLogGroups[group] = true
		}
	}

	// Get all log groups
	var logGroups []string
	err = cwLogs.DescribeLogGroupsPages(&cloudwatchlogs.DescribeLogGroupsInput{},
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			for _, group := range page.LogGroups {
				logGroups = append(logGroups, *group.LogGroupName)
			}
			return !lastPage
		})

	if err != nil {
		fmt.Printf("Error fetching log groups: %v\n", err)
		os.Exit(1)
	}

	// Initialize metrics
	metrics := make([]LogGroupMetrics, 0, len(logGroups))

	// Process each log group
	for _, logGroup := range logGroups {
		// Get ingested bytes metric for this log group
		ingestedBytes := getMetricStatistics(cw, logGroup, "IncomingBytes", firstDay, lastDay)

		// Get stored bytes (only if detailed view is enabled)
		var storedBytes float64
		if detailedView {
			storedBytes = getMetricStatistics(cw, logGroup, "LogGroupSizeBytes", firstDay, lastDay)
			// If no data point for the period, try to get the latest datapoint
			if storedBytes == 0 {
				storedBytes = getLatestMetricDatapoint(cw, logGroup, "LogGroupSizeBytes")
			}
		}

		// Skip groups that are smaller than the minimum size
		if ingestedBytes/1048576 < minSizeMB {
			continue
		}

		// Check if this is an app log group
		isAppLog := appLogGroups[logGroup]

		metrics = append(metrics, LogGroupMetrics{
			Name:          logGroup,
			IngestedBytes: ingestedBytes,
			StoredBytes:   storedBytes,
			IsAppLog:      isAppLog,
		})
	}

	// Sort metrics by size if requested
	if sortBySize {
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].IngestedBytes > metrics[j].IngestedBytes
		})
	}

	// Output results
	displayResults(metrics, detailedView)

	// Write CSV if requested
	if outputCSV != "" {
		writeCSV(outputCSV, metrics, detailedView)
	}
}

// getMetricStatistics retrieves CloudWatch metric statistics for a log group
func getMetricStatistics(cw *cloudwatch.CloudWatch, logGroup, metricName string, startTime, endTime time.Time) float64 {
	// Calculate period (must be multiple of 60 seconds)
	totalSeconds := int64(endTime.Sub(startTime).Seconds())
	// Round up to nearest multiple of 60
	period := ((totalSeconds + 59) / 60) * 60
	// Ensure minimum of 60 seconds
	if period < 60 {
		period = 60
	}
	
	params := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/Logs"),
		MetricName: aws.String(metricName),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("LogGroupName"),
				Value: aws.String(logGroup),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int64(period),
		Statistics: []*string{aws.String("Sum")},
	}

	// For StorageBytes, use Average instead of Sum
	if metricName == "LogGroupSizeBytes" {
		params.Statistics = []*string{aws.String("Average")}
	}

	resp, err := cw.GetMetricStatistics(params)
	if err != nil {
		fmt.Printf("Error getting %s metrics for %s: %v\n", metricName, logGroup, err)
		return 0
	}

	if len(resp.Datapoints) > 0 {
		if metricName == "LogGroupSizeBytes" {
			return *resp.Datapoints[0].Average
		}
		return *resp.Datapoints[0].Sum
	}

	return 0
}

// getLatestMetricDatapoint gets the most recent datapoint for a metric
func getLatestMetricDatapoint(cw *cloudwatch.CloudWatch, logGroup, metricName string) float64 {
	// Get a datapoint from the last 24 hours
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	params := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/Logs"),
		MetricName: aws.String(metricName),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("LogGroupName"),
				Value: aws.String(logGroup),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int64(3600), // 1 hour periods (already a multiple of 60)
		Statistics: []*string{aws.String("Average")},
	}

	resp, err := cw.GetMetricStatistics(params)
	if err != nil {
		fmt.Printf("Error getting latest %s metrics for %s: %v\n", metricName, logGroup, err)
		return 0
	}

	if len(resp.Datapoints) > 0 {
		// Find the most recent datapoint
		var latestDatapoint *cloudwatch.Datapoint
		for _, dp := range resp.Datapoints {
			if latestDatapoint == nil || dp.Timestamp.After(*latestDatapoint.Timestamp) {
				latestDatapoint = dp
			}
		}
		return *latestDatapoint.Average
	}

	return 0
}

// displayResults outputs the results to the console
func displayResults(metrics []LogGroupMetrics, detailed bool) {
	// Calculate totals
	var totalIngestedBytes, appTotalIngestedBytes float64
	var totalStoredBytes, appTotalStoredBytes float64

	// Display individual log group info
	fmt.Println("\nLog Group Metrics:")
	fmt.Println("----------------------------------------")

	for _, m := range metrics {
		ingestedMB := m.IngestedBytes / 1048576

		if detailed {
			storedMB := m.StoredBytes / 1048576
			fmt.Printf("%-60s Ingested: %9.2f MB  Stored: %9.2f MB", m.Name, ingestedMB, storedMB)
			if m.IsAppLog {
				fmt.Print("  [APP]")
			}
			fmt.Println()
		} else {
			fmt.Printf("%-60s Size: %9.2f MB", m.Name, ingestedMB)
			if m.IsAppLog {
				fmt.Print("  [APP]")
			}
			fmt.Println()
		}

		totalIngestedBytes += m.IngestedBytes
		totalStoredBytes += m.StoredBytes

		if m.IsAppLog {
			appTotalIngestedBytes += m.IngestedBytes
			appTotalStoredBytes += m.StoredBytes
		}
	}

	// Convert totals to MB and GB for readability
	totalIngestedMB := totalIngestedBytes / 1048576
	totalIngestedGB := totalIngestedBytes / 1073741824
	appTotalIngestedMB := appTotalIngestedBytes / 1048576
	appTotalIngestedGB := appTotalIngestedBytes / 1073741824

	// Display summary
	fmt.Println("\nSummary:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Total log groups: %d\n", len(metrics))
	fmt.Printf("Total ingested: %.2f MB (%.2f GB)\n", totalIngestedMB, totalIngestedGB)
	fmt.Printf("App logs ingested: %.2f MB (%.2f GB)\n", appTotalIngestedMB, appTotalIngestedGB)

	if detailed {
		totalStoredMB := totalStoredBytes / 1048576
		totalStoredGB := totalStoredBytes / 1073741824
		appTotalStoredMB := appTotalStoredBytes / 1048576
		appTotalStoredGB := appTotalStoredBytes / 1073741824

		// Approximate costs (using $0.50 per GB for ingestion, $0.03 per GB for storage)
		ingestCost := totalIngestedGB * 0.50
		storageCost := totalStoredGB * 0.03
		appIngestCost := appTotalIngestedGB * 0.50
		appStorageCost := appTotalStoredGB * 0.03

		fmt.Printf("Total stored: %.2f MB (%.2f GB)\n", totalStoredMB, totalStoredGB)
		fmt.Printf("App logs stored: %.2f MB (%.2f GB)\n", appTotalStoredMB, appTotalStoredGB)
		fmt.Printf("\nEstimated Costs:\n")
		fmt.Printf("Total ingestion cost: $%.2f\n", ingestCost)
		fmt.Printf("Total storage cost: $%.2f\n", storageCost)
		fmt.Printf("App logs ingestion cost: $%.2f\n", appIngestCost)
		fmt.Printf("App logs storage cost: $%.2f\n", appStorageCost)
	}
}

// writeCSV outputs the results to a CSV file
func writeCSV(filename string, metrics []LogGroupMetrics, detailed bool) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer file.Close()

	// Write header
	if detailed {
		file.WriteString("LogGroup,IngestedBytes,IngestedMB,StoredBytes,StoredMB,IsAppLog\n")
	} else {
		file.WriteString("LogGroup,IngestedBytes,IngestedMB,IsAppLog\n")
	}

	// Write data
	for _, m := range metrics {
		ingestedMB := m.IngestedBytes / 1048576
		isApp := "false"
		if m.IsAppLog {
			isApp = "true"
		}

		if detailed {
			storedMB := m.StoredBytes / 1048576
			file.WriteString(fmt.Sprintf("%s,%.0f,%.2f,%.0f,%.2f,%s\n",
				m.Name, m.IngestedBytes, ingestedMB, m.StoredBytes, storedMB, isApp))
		} else {
			file.WriteString(fmt.Sprintf("%s,%.0f,%.2f,%s\n",
				m.Name, m.IngestedBytes, ingestedMB, isApp))
		}
	}

	fmt.Printf("\nResults written to %s\n", filename)
}
