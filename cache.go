package ecrauditor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	dir string
}

func NewCache() (*Cache, error) {
	cacheDir, err := defaultCacheDir()
	if err != nil {
		return nil, err
	}
	return &Cache{dir: cacheDir}, nil
}

func NewCacheWithDir(dir string) *Cache {
	return &Cache{dir: dir}
}

func defaultCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".ecr-auditor", "cache"), nil
}

func (c *Cache) cacheFilePath(region, profile string) string {
	filename := fmt.Sprintf("%s_%s.json", profile, region)
	return filepath.Join(c.dir, filename)
}

func (c *Cache) Save(result *AuditResult) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal audit result: %w", err)
	}

	filePath := c.cacheFilePath(result.Region, result.Profile)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

func (c *Cache) Load(region, profile string) (*AuditResult, error) {
	filePath := c.cacheFilePath(region, profile)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheNotFound
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var result AuditResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &result, nil
}

func (c *Cache) Exists(region, profile string) bool {
	filePath := c.cacheFilePath(region, profile)
	_, err := os.Stat(filePath)
	return err == nil
}

func (c *Cache) Clear(region, profile string) error {
	filePath := c.cacheFilePath(region, profile)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	return nil
}
