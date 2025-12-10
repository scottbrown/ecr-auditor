package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	ecrauditor "github.com/scottbrown/ecr-auditor"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Display audit results summary",
	Long:  `Display a summary of the cached ECR audit results.`,
	RunE:  runSummary,
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
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

	printSummary(result)
	return nil
}

func printSummary(result *ecrauditor.AuditResult) {
	fmt.Println("ECR Audit Summary")
	fmt.Println("=================")
	fmt.Printf("Region:      %s\n", result.Region)
	fmt.Printf("Profile:     %s\n", result.Profile)
	fmt.Printf("Collected:   %s\n", result.CollectedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Println()

	fmt.Println("Overview")
	fmt.Println("--------")
	fmt.Printf("Total Repositories:           %d\n", result.RepositoryCount)
	fmt.Printf("Total Images:                 %d\n", result.TotalImageCount())
	fmt.Printf("With Lifecycle Policy:        %d\n", result.RepositoriesWithLifecyclePolicy())
	fmt.Printf("Without Lifecycle Policy:     %d\n", result.RepositoriesWithoutLifecyclePolicy())
	fmt.Println()

	fmt.Println("Repositories")
	fmt.Println("------------")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tIMAGES\tMEDIAN/DAY\tLIFECYCLE POLICY\tCREATED")
	for _, repo := range result.Repositories {
		policyStatus := "No"
		if repo.LifecyclePolicy != nil {
			policyStatus = "Yes"
		}
		fmt.Fprintf(w, "%s\t%d\t%.1f\t%s\t%s\n",
			repo.Name,
			repo.ImageCount,
			repo.MedianImagesPerDay(),
			policyStatus,
			repo.CreatedAt.Format("2006-01-02"),
		)
	}
	w.Flush()

	if result.RepositoriesWithoutLifecyclePolicy() > 0 {
		fmt.Println()
		fmt.Println("Repositories Missing Lifecycle Policies")
		fmt.Println("----------------------------------------")
		for _, repo := range result.Repositories {
			if repo.LifecyclePolicy == nil {
				fmt.Printf("  - %s (%d images)\n", repo.Name, repo.ImageCount)
			}
		}
	}
}
