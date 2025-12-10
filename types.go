package ecrauditor

import (
	"sort"
	"time"
)

type LifecyclePolicy struct {
	RegistryID     string `json:"registry_id"`
	RepositoryName string `json:"repository_name"`
	PolicyText     string `json:"policy_text"`
}

type Image struct {
	Digest          string    `json:"digest"`
	Tags            []string  `json:"tags"`
	PushedAt        time.Time `json:"pushed_at"`
	SizeBytes       int64     `json:"size_bytes"`
	ImageScanStatus string    `json:"image_scan_status,omitempty"`
}

type Repository struct {
	Name            string           `json:"name"`
	URI             string           `json:"uri"`
	CreatedAt       time.Time        `json:"created_at"`
	ImageCount      int              `json:"image_count"`
	Images          []Image          `json:"images"`
	LifecyclePolicy *LifecyclePolicy `json:"lifecycle_policy,omitempty"`
}

type AuditResult struct {
	Region          string       `json:"region"`
	Profile         string       `json:"profile"`
	CollectedAt     time.Time    `json:"collected_at"`
	RepositoryCount int          `json:"repository_count"`
	Repositories    []Repository `json:"repositories"`
}

func (a *AuditResult) TotalImageCount() int {
	total := 0
	for _, repo := range a.Repositories {
		total += repo.ImageCount
	}
	return total
}

func (a *AuditResult) RepositoriesWithLifecyclePolicy() int {
	count := 0
	for _, repo := range a.Repositories {
		if repo.LifecyclePolicy != nil {
			count++
		}
	}
	return count
}

func (a *AuditResult) RepositoriesWithoutLifecyclePolicy() int {
	return a.RepositoryCount - a.RepositoriesWithLifecyclePolicy()
}

func (r *Repository) MedianImagesPerDay() float64 {
	if len(r.Images) == 0 {
		return 0
	}

	dailyCounts := make(map[string]int)
	for _, img := range r.Images {
		day := img.PushedAt.Format("2006-01-02")
		dailyCounts[day]++
	}

	if len(dailyCounts) == 0 {
		return 0
	}

	counts := make([]int, 0, len(dailyCounts))
	for _, count := range dailyCounts {
		counts = append(counts, count)
	}

	sort.Ints(counts)

	n := len(counts)
	if n%2 == 0 {
		return float64(counts[n/2-1]+counts[n/2]) / 2.0
	}
	return float64(counts[n/2])
}
