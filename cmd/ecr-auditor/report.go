package main

import (
	"errors"
	"fmt"
	"os"

	ecrauditor "github.com/scottbrown/ecr-auditor"
	"github.com/spf13/cobra"
)

var outputFile string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate an HTML report",
	Long:  `Generate a static, interactive HTML report of the ECR audit results.`,
	RunE:  runReport,
}

func init() {
	reportCmd.Flags().StringVarP(&outputFile, "output", "o", "report.html", "Output file path")
	rootCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, args []string) error {
	cache, err := ecrauditor.NewCache()
	if err != nil {
		return fmt.Errorf("failed to initialise cache: %w", err)
	}

	result, err := cache.Load(region, profile)
	if err != nil {
		if errors.Is(err, ecrauditor.ErrCacheNotFound) {
			return fmt.Errorf("no cached data found. Run 'ecr-auditor collect --region %s --profile %s' first", region, profile)
		}
		return fmt.Errorf("failed to load cache: %w", err)
	}

	generator, err := ecrauditor.NewReportGenerator()
	if err != nil {
		return fmt.Errorf("failed to create report generator: %w", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := generator.Generate(file, result); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("Report generated: %s\n", outputFile)
	return nil
}
