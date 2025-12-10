package ecrauditor

import (
	"testing"
	"time"
)

func TestAuditResult_TotalImageCount(t *testing.T) {
	tests := []struct {
		name     string
		result   AuditResult
		expected int
	}{
		{
			name: "empty repositories",
			result: AuditResult{
				Repositories: []Repository{},
			},
			expected: 0,
		},
		{
			name: "single repository with images",
			result: AuditResult{
				Repositories: []Repository{
					{Name: "repo1", ImageCount: 5},
				},
			},
			expected: 5,
		},
		{
			name: "multiple repositories",
			result: AuditResult{
				Repositories: []Repository{
					{Name: "repo1", ImageCount: 5},
					{Name: "repo2", ImageCount: 10},
					{Name: "repo3", ImageCount: 3},
				},
			},
			expected: 18,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.result.TotalImageCount()
			if actual != tt.expected {
				t.Errorf("TotalImageCount() = %d, expected %d", actual, tt.expected)
			}
		})
	}
}

func TestAuditResult_RepositoriesWithLifecyclePolicy(t *testing.T) {
	tests := []struct {
		name     string
		result   AuditResult
		expected int
	}{
		{
			name: "no repositories",
			result: AuditResult{
				Repositories: []Repository{},
			},
			expected: 0,
		},
		{
			name: "no policies",
			result: AuditResult{
				Repositories: []Repository{
					{Name: "repo1"},
					{Name: "repo2"},
				},
			},
			expected: 0,
		},
		{
			name: "all have policies",
			result: AuditResult{
				Repositories: []Repository{
					{Name: "repo1", LifecyclePolicy: &LifecyclePolicy{}},
					{Name: "repo2", LifecyclePolicy: &LifecyclePolicy{}},
				},
			},
			expected: 2,
		},
		{
			name: "mixed",
			result: AuditResult{
				Repositories: []Repository{
					{Name: "repo1", LifecyclePolicy: &LifecyclePolicy{}},
					{Name: "repo2"},
					{Name: "repo3", LifecyclePolicy: &LifecyclePolicy{}},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.result.RepositoriesWithLifecyclePolicy()
			if actual != tt.expected {
				t.Errorf("RepositoriesWithLifecyclePolicy() = %d, expected %d", actual, tt.expected)
			}
		})
	}
}

func TestAuditResult_RepositoriesWithoutLifecyclePolicy(t *testing.T) {
	tests := []struct {
		name     string
		result   AuditResult
		expected int
	}{
		{
			name: "no repositories",
			result: AuditResult{
				RepositoryCount: 0,
				Repositories:    []Repository{},
			},
			expected: 0,
		},
		{
			name: "all without policies",
			result: AuditResult{
				RepositoryCount: 2,
				Repositories: []Repository{
					{Name: "repo1"},
					{Name: "repo2"},
				},
			},
			expected: 2,
		},
		{
			name: "mixed",
			result: AuditResult{
				RepositoryCount: 3,
				Repositories: []Repository{
					{Name: "repo1", LifecyclePolicy: &LifecyclePolicy{}},
					{Name: "repo2"},
					{Name: "repo3"},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.result.RepositoriesWithoutLifecyclePolicy()
			if actual != tt.expected {
				t.Errorf("RepositoriesWithoutLifecyclePolicy() = %d, expected %d", actual, tt.expected)
			}
		})
	}
}

func TestRepository_Fields(t *testing.T) {
	now := time.Now()
	repo := Repository{
		Name:       "test-repo",
		URI:        "123456789.dkr.ecr.us-east-1.amazonaws.com/test-repo",
		CreatedAt:  now,
		ImageCount: 5,
		Images: []Image{
			{Digest: "sha256:abc123", Tags: []string{"latest", "v1.0"}},
		},
		LifecyclePolicy: &LifecyclePolicy{
			RegistryID:     "123456789",
			RepositoryName: "test-repo",
			PolicyText:     `{"rules":[]}`,
		},
	}

	if repo.Name != "test-repo" {
		t.Errorf("Name = %s, expected test-repo", repo.Name)
	}
	if repo.ImageCount != 5 {
		t.Errorf("ImageCount = %d, expected 5", repo.ImageCount)
	}
	if len(repo.Images) != 1 {
		t.Errorf("Images length = %d, expected 1", len(repo.Images))
	}
	if repo.LifecyclePolicy == nil {
		t.Error("LifecyclePolicy should not be nil")
	}
}

func TestImage_Fields(t *testing.T) {
	now := time.Now()
	img := Image{
		Digest:          "sha256:abc123",
		Tags:            []string{"latest", "v1.0"},
		PushedAt:        now,
		SizeBytes:       1024000,
		ImageScanStatus: "COMPLETE",
	}

	if img.Digest != "sha256:abc123" {
		t.Errorf("Digest = %s, expected sha256:abc123", img.Digest)
	}
	if len(img.Tags) != 2 {
		t.Errorf("Tags length = %d, expected 2", len(img.Tags))
	}
	if img.SizeBytes != 1024000 {
		t.Errorf("SizeBytes = %d, expected 1024000", img.SizeBytes)
	}
}

func TestRepository_MedianImagesPerDay(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		repo     Repository
		expected float64
	}{
		{
			name:     "no images",
			repo:     Repository{Images: []Image{}},
			expected: 0,
		},
		{
			name: "single image on one day",
			repo: Repository{
				Images: []Image{
					{PushedAt: baseTime},
				},
			},
			expected: 1,
		},
		{
			name: "multiple images on same day",
			repo: Repository{
				Images: []Image{
					{PushedAt: baseTime},
					{PushedAt: baseTime.Add(time.Hour)},
					{PushedAt: baseTime.Add(2 * time.Hour)},
				},
			},
			expected: 3,
		},
		{
			name: "one image per day across multiple days",
			repo: Repository{
				Images: []Image{
					{PushedAt: baseTime},
					{PushedAt: baseTime.AddDate(0, 0, 1)},
					{PushedAt: baseTime.AddDate(0, 0, 2)},
				},
			},
			expected: 1,
		},
		{
			name: "varying images per day - odd count",
			repo: Repository{
				Images: []Image{
					{PushedAt: baseTime},
					{PushedAt: baseTime.Add(time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 1)},
					{PushedAt: baseTime.AddDate(0, 0, 1).Add(time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 1).Add(2 * time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 2)},
				},
			},
			expected: 2,
		},
		{
			name: "varying images per day - even count for median average",
			repo: Repository{
				Images: []Image{
					{PushedAt: baseTime},
					{PushedAt: baseTime.AddDate(0, 0, 1)},
					{PushedAt: baseTime.AddDate(0, 0, 1).Add(time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 1).Add(2 * time.Hour)},
				},
			},
			expected: 2,
		},
		{
			name: "median with even day count",
			repo: Repository{
				Images: []Image{
					{PushedAt: baseTime},
					{PushedAt: baseTime.AddDate(0, 0, 1)},
					{PushedAt: baseTime.AddDate(0, 0, 1).Add(time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 2)},
					{PushedAt: baseTime.AddDate(0, 0, 2).Add(time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 2).Add(2 * time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 3)},
					{PushedAt: baseTime.AddDate(0, 0, 3).Add(time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 3).Add(2 * time.Hour)},
					{PushedAt: baseTime.AddDate(0, 0, 3).Add(3 * time.Hour)},
				},
			},
			expected: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.repo.MedianImagesPerDay()
			if actual != tt.expected {
				t.Errorf("MedianImagesPerDay() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}
