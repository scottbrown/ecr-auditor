package main

import (
	"context"
	"fmt"
	"os"

	ecrauditor "github.com/scottbrown/ecr-auditor"
	"github.com/spf13/cobra"
)

var forceRefresh bool

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect ECR audit data",
	Long: `Collect audit data from AWS ECR, including repositories, images,
and lifecycle policies. Data is cached locally for subsequent queries.
Use --force to refresh the cache.`,
	RunE: runCollect,
}

func init() {
	collectCmd.Flags().BoolVarP(&forceRefresh, "force", "f", false, "Force refresh of cached data")
	rootCmd.AddCommand(collectCmd)
}

func runCollect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cache, err := ecrauditor.NewCache()
	if err != nil {
		return fmt.Errorf("failed to initialise cache: %w", err)
	}

	if !forceRefresh && cache.Exists(region, profile) {
		fmt.Println("Cache exists. Use --force to refresh.")
		return nil
	}

	fmt.Printf("Collecting ECR data for region %s with profile %s...\n", region, profile)

	auditor, err := ecrauditor.NewAuditor(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("failed to create auditor: %w", err)
	}

	result, err := auditor.Collect(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect audit data: %w", err)
	}

	if err := cache.Save(result); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	fmt.Printf("Collected data for %d repositories with %d total images.\n",
		result.RepositoryCount, result.TotalImageCount())
	fmt.Printf("Data cached at: %s\n", cacheLocation(region, profile))

	return nil
}

func cacheLocation(region, profile string) string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/.ecr-auditor/cache/%s_%s.json", homeDir, profile, region)
}
