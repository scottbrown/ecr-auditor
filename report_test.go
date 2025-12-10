package ecrauditor

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewReportGenerator(t *testing.T) {
	gen, err := NewReportGenerator()
	if err != nil {
		t.Fatalf("NewReportGenerator() error = %v", err)
	}
	if gen == nil {
		t.Error("NewReportGenerator() returned nil")
	}
}

func TestReportGenerator_Generate(t *testing.T) {
	gen, err := NewReportGenerator()
	if err != nil {
		t.Fatalf("NewReportGenerator() error = %v", err)
	}

	result := &AuditResult{
		Region:          "us-east-1",
		Profile:         "test-profile",
		CollectedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		RepositoryCount: 2,
		Repositories: []Repository{
			{
				Name:       "my-app",
				URI:        "123456789.dkr.ecr.us-east-1.amazonaws.com/my-app",
				CreatedAt:  time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
				ImageCount: 2,
				Images: []Image{
					{
						Digest:    "sha256:abc123",
						Tags:      []string{"latest", "v1.0"},
						PushedAt:  time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						SizeBytes: 104857600,
					},
					{
						Digest:    "sha256:def456",
						PushedAt:  time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
						SizeBytes: 52428800,
					},
				},
				LifecyclePolicy: &LifecyclePolicy{
					RegistryID:     "123456789",
					RepositoryName: "my-app",
					PolicyText:     `{"rules":[]}`,
				},
			},
			{
				Name:       "another-app",
				URI:        "123456789.dkr.ecr.us-east-1.amazonaws.com/another-app",
				CreatedAt:  time.Date(2023, 8, 15, 0, 0, 0, 0, time.UTC),
				ImageCount: 0,
				Images:     []Image{},
			},
		},
	}

	var buf bytes.Buffer
	err = gen.Generate(&buf, result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	html := buf.String()

	expectedStrings := []string{
		"<!DOCTYPE html>",
		"ECR Audit Report",
		"us-east-1",
		"test-profile",
		"my-app",
		"another-app",
		"2024-01-15",
		`<span class="badge success">Yes</span>`,
		`<span class="badge danger">No</span>`,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(html, expected) {
			t.Errorf("Generated HTML should contain %q", expected)
		}
	}
}

func TestReportGenerator_Generate_EmptyResult(t *testing.T) {
	gen, err := NewReportGenerator()
	if err != nil {
		t.Fatalf("NewReportGenerator() error = %v", err)
	}

	result := &AuditResult{
		Region:          "eu-west-1",
		Profile:         "prod",
		CollectedAt:     time.Now(),
		RepositoryCount: 0,
		Repositories:    []Repository{},
	}

	var buf bytes.Buffer
	err = gen.Generate(&buf, result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "eu-west-1") {
		t.Error("Generated HTML should contain region")
	}
	if !strings.Contains(html, "prod") {
		t.Error("Generated HTML should contain profile")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
		{104857600, "100.00 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			actual := formatBytes(tt.bytes)
			if actual != tt.expected {
				t.Errorf("formatBytes(%d) = %s, expected %s", tt.bytes, actual, tt.expected)
			}
		})
	}
}

func TestReportGenerator_Generate_WithImageTags(t *testing.T) {
	gen, err := NewReportGenerator()
	if err != nil {
		t.Fatalf("NewReportGenerator() error = %v", err)
	}

	result := &AuditResult{
		Region:          "us-east-1",
		Profile:         "test",
		CollectedAt:     time.Now(),
		RepositoryCount: 1,
		Repositories: []Repository{
			{
				Name:       "tagged-repo",
				CreatedAt:  time.Now(),
				ImageCount: 1,
				Images: []Image{
					{
						Digest:    "sha256:abc",
						Tags:      []string{"production", "v2.0.0", "stable"},
						PushedAt:  time.Now(),
						SizeBytes: 1024,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err = gen.Generate(&buf, result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	html := buf.String()
	for _, tag := range []string{"production", "v2.0.0", "stable"} {
		if !strings.Contains(html, tag) {
			t.Errorf("Generated HTML should contain tag %q", tag)
		}
	}
}

func TestReportGenerator_Generate_UntaggedImages(t *testing.T) {
	gen, err := NewReportGenerator()
	if err != nil {
		t.Fatalf("NewReportGenerator() error = %v", err)
	}

	result := &AuditResult{
		Region:          "us-east-1",
		Profile:         "test",
		CollectedAt:     time.Now(),
		RepositoryCount: 1,
		Repositories: []Repository{
			{
				Name:       "untagged-repo",
				CreatedAt:  time.Now(),
				ImageCount: 1,
				Images: []Image{
					{
						Digest:    "sha256:abc",
						Tags:      nil,
						PushedAt:  time.Now(),
						SizeBytes: 1024,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err = gen.Generate(&buf, result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "untagged") {
		t.Error("Generated HTML should indicate untagged images")
	}
}
