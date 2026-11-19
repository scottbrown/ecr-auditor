# ecr-auditor

A CLI tool for auditing AWS ECR (Elastic Container Registry) repositories. It collects information about private repositories, their images, and lifecycle policies, caching results for efficient subsequent queries.

## Features

- Audit ECR repositories in any AWS region
- Collect image counts and metadata per repository
- Track repository growth with median images per day metric
- Check lifecycle policy compliance
- Cache audit results locally for fast subsequent queries
- Generate interactive HTML reports
- Display summary statistics in the terminal

## Installation

### Prerequisites

- Go 1.27.4 or later
- [Task](https://taskfile.dev/) (optional, for build automation)
- AWS credentials configured (via `~/.aws/credentials` or environment variables)

### Building from source

```bash
# Clone the repository
git clone https://github.com/scottbrown/ecr-auditor.git
cd ecr-auditor

# Build using Task
task build

# Or build directly with Go
go build -o bin/ecr-auditor ./cmd/ecr-auditor
```

### Installing

```bash
task install
# Or manually copy to your PATH
cp bin/ecr-auditor /usr/local/bin/
```

## Usage

### Collecting audit data

Collect ECR data from a specific region:

```bash
ecr-auditor collect --region us-east-1 --profile my-aws-profile
```

Options:
- `--region` (required): AWS region to audit
- `--profile`: AWS profile name (default: "default")
- `--force`, `-f`: Force refresh of cached data

### Viewing summary

Display a summary of the cached audit results:

```bash
ecr-auditor summary --region us-east-1 --profile my-aws-profile
```

Example output:

```
ECR Audit Summary
=================
Region:      us-east-1
Profile:     my-aws-profile
Collected:   2024-01-15 10:30:00 UTC

Overview
--------
Total Repositories:           25
Total Images:                 142
With Lifecycle Policy:        18
Without Lifecycle Policy:     7

Repositories
------------
NAME                    IMAGES  MEDIAN/DAY  LIFECYCLE POLICY  CREATED
my-app                  12      2.0         Yes               2023-06-01
another-service         8       1.5         No                2023-08-15
...
```

### Generating HTML reports

Generate an interactive HTML report:

```bash
ecr-auditor report --region us-east-1 --profile my-aws-profile
```

Options:
- `--output`, `-o`: Output file path (default: "report.html")

```bash
ecr-auditor report --region us-east-1 --output ecr-audit-report.html
```

The generated HTML report includes:
- Summary statistics cards
- Filterable repository table with median images per day
- Expandable image details
- Lifecycle policy compliance indicators
- Responsive design for viewing on any device

## Cache

Audit data is cached locally at `~/.ecr-auditor/cache/`. The cache file is named `{profile}_{region}.json`.

To refresh the cache:

```bash
ecr-auditor collect --region us-east-1 --force
```

## Development

### Running tests

```bash
task test

# With coverage
task test:coverage

# Generate HTML coverage report
task test:coverage:html
```

### Available tasks

```bash
task --list
```

- `task build` - Build the binary
- `task test` - Run tests
- `task test:coverage` - Run tests with coverage
- `task lint` - Run linter
- `task fmt` - Format code
- `task vet` - Run go vet
- `task clean` - Clean build artifacts
- `task install` - Install the binary
- `task all` - Run fmt, vet, test, and build

## AWS Permissions

The tool requires the following IAM permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:DescribeRepositories",
                "ecr:DescribeImages",
                "ecr:GetLifecyclePolicy"
            ],
            "Resource": "*"
        }
    ]
}
```

## Licence

[MIT](LICENSE)
