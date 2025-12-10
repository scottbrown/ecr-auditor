package ecrauditor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	result := &AuditResult{
		Region:          "us-east-1",
		Profile:         "test-profile",
		CollectedAt:     time.Now().Truncate(time.Second),
		RepositoryCount: 2,
		Repositories: []Repository{
			{
				Name:       "repo1",
				URI:        "123456789.dkr.ecr.us-east-1.amazonaws.com/repo1",
				CreatedAt:  time.Now().Truncate(time.Second),
				ImageCount: 3,
			},
			{
				Name:       "repo2",
				URI:        "123456789.dkr.ecr.us-east-1.amazonaws.com/repo2",
				CreatedAt:  time.Now().Truncate(time.Second),
				ImageCount: 5,
				LifecyclePolicy: &LifecyclePolicy{
					RegistryID:     "123456789",
					RepositoryName: "repo2",
					PolicyText:     `{"rules":[]}`,
				},
			},
		},
	}

	err := cache.Save(result)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := cache.Load("us-east-1", "test-profile")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Region != result.Region {
		t.Errorf("Region = %s, expected %s", loaded.Region, result.Region)
	}
	if loaded.Profile != result.Profile {
		t.Errorf("Profile = %s, expected %s", loaded.Profile, result.Profile)
	}
	if loaded.RepositoryCount != result.RepositoryCount {
		t.Errorf("RepositoryCount = %d, expected %d", loaded.RepositoryCount, result.RepositoryCount)
	}
	if len(loaded.Repositories) != len(result.Repositories) {
		t.Errorf("Repositories length = %d, expected %d", len(loaded.Repositories), len(result.Repositories))
	}
}

func TestCache_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	_, err := cache.Load("us-east-1", "nonexistent")
	if !errors.Is(err, ErrCacheNotFound) {
		t.Errorf("Load() error = %v, expected ErrCacheNotFound", err)
	}
}

func TestCache_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	if cache.Exists("us-east-1", "test") {
		t.Error("Exists() = true, expected false for nonexistent cache")
	}

	result := &AuditResult{
		Region:  "us-east-1",
		Profile: "test",
	}
	if err := cache.Save(result); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !cache.Exists("us-east-1", "test") {
		t.Error("Exists() = false, expected true after save")
	}
}

func TestCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	result := &AuditResult{
		Region:  "us-east-1",
		Profile: "test",
	}
	if err := cache.Save(result); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !cache.Exists("us-east-1", "test") {
		t.Fatal("Cache should exist after save")
	}

	if err := cache.Clear("us-east-1", "test"); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if cache.Exists("us-east-1", "test") {
		t.Error("Cache should not exist after clear")
	}
}

func TestCache_Clear_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	err := cache.Clear("us-east-1", "nonexistent")
	if err != nil {
		t.Errorf("Clear() should not error for nonexistent file, got %v", err)
	}
}

func TestCache_cacheFilePath(t *testing.T) {
	cache := NewCacheWithDir("/tmp/test-cache")

	path := cache.cacheFilePath("us-east-1", "myprofile")
	expected := "/tmp/test-cache/myprofile_us-east-1.json"

	if path != expected {
		t.Errorf("cacheFilePath() = %s, expected %s", path, expected)
	}
}

func TestNewCache(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".ecr-auditor", "cache")

	if cache.dir != expectedDir {
		t.Errorf("cache.dir = %s, expected %s", cache.dir, expectedDir)
	}
}

func TestCache_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "nested", "cache", "dir")
	cache := NewCacheWithDir(cacheDir)

	result := &AuditResult{
		Region:  "us-east-1",
		Profile: "test",
	}

	if err := cache.Save(result); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Save() should create cache directory")
	}
}
