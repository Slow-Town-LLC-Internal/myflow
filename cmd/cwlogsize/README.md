# CloudWatch Log Size Calculator

A Go tool for calculating AWS CloudWatch log sizes with support for filtering, sorting, and detailed cost analysis.

## Features

- Calculate log sizes for a specific month period
- Filter and track application logs separately
- Sort log groups by size
- Generate detailed metrics including storage costs
- Export results to CSV
- Works across all platforms (macOS, Linux, Windows)
- Supports all AWS authentication methods including SSO

## Setup Instructions

1. First, create a new directory for the project:
   ```bash
   mkdir cwlogsize
   cd cwlogsize
   ```

2. Initialize the Go module:
   ```bash
   go mod init cwlogsize
   ```

3. Install the required dependencies:
   ```bash
   go get github.com/aws/aws-sdk-go/aws
   go get github.com/aws/aws-sdk-go/aws/session
   go get github.com/aws/aws-sdk-go/service/cloudwatch
   go get github.com/aws/aws-sdk-go/service/cloudwatchlogs
   ```

4. Save the code to a file named `main.go`

5. Build the binary:
   ```bash
   go build -o cwlogsize .
   ```

## Usage Examples

Basic usage to get last month's log sizes:
```bash
./cwlogsize
```

Specify your AWS profile and region:
```bash
./cwlogsize --aws-profile=your-profile --aws-region=us-west-2
```

Track specific app log groups separately:
```bash
./cwlogsize --log-groups=/aws/lambda/app1,/aws/lambda/app2,/aws/lambda/app3
```

Get logs from 3 months ago:
```bash
./cwlogsize --months-back=3
```

Sort log groups by size (largest first):
```bash
./cwlogsize --sort
```

Only show log groups larger than 10MB:
```bash
./cwlogsize --min-size=10
```

Show detailed metrics including storage size and estimated costs:
```bash
./cwlogsize --detailed
```

Export results to CSV:
```bash
./cwlogsize --sort --detailed --csv=logs.csv
```

## Command Line Options

| Option | Description |
|--------|-------------|
| `--aws-profile` | AWS profile to use (default: default profile) |
| `--aws-region` | AWS region (default: us-east-1) |
| `--log-groups` | Comma-separated list of app log groups to track separately |
| `--months-back` | Number of months back to calculate (default: 1) |
| `--sort` | Sort log groups by size (largest first) |
| `--min-size` | Only show log groups larger than this size (in MB) |
| `--detailed` | Show detailed metrics including storage costs |
| `--csv` | Output results to CSV file (optional) |

## Why Use This Tool

This tool provides several advantages over bash scripts:

1. **Platform independence** - works on macOS, Linux, and Windows without any modifications
2. **Better error handling** - properly handles AWS API errors and edge cases
3. **More features** - sorting, filtering, detailed metrics, and CSV export
4. **Better performance** - uses the CloudWatch metrics API efficiently

The tool will automatically authenticate using your AWS credentials in the same way the AWS CLI does, supporting SSO and all other authentication methods.
