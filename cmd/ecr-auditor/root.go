package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	region  string
	profile string
)

var rootCmd = &cobra.Command{
	Use:   "ecr-auditor",
	Short: "Audit AWS ECR repositories",
	Long: `ecr-auditor is a CLI tool that audits AWS ECR repositories.
It collects information about private repositories, their images,
and lifecycle policies, caching the results for subsequent queries.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region (required)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "default", "AWS profile name")

	rootCmd.MarkPersistentFlagRequired("region")
}
