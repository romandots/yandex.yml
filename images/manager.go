package images

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"yandex-export/config"
)

// ImageManager handles random image selection with usage tracking
type ImageManager struct {
	mu           sync.RWMutex
	usageStats   map[string]map[string]int // categoryID -> imagePath -> usage count
	imageCache   map[string][]string       // categoryID -> []imagePaths
	lastScanTime map[string]time.Time      // categoryID -> last scan time
}

// NewImageManager creates a new image manager instance
func NewImageManager() *ImageManager {
	return &ImageManager{
		usageStats:   make(map[string]map[string]int),
		imageCache:   make(map[string][]string),
		lastScanTime: make(map[string]time.Time),
	}
}

// GetRandomImage returns a random image path for the given category ID
// It prioritizes images that have been used less frequently
func (im *ImageManager) GetRandomImage(categoryID int) (string, error) {
	categoryStr := fmt.Sprintf("%d", categoryID)

	im.mu.Lock()
	defer im.mu.Unlock()

	// Refresh image cache if needed (scan directory every 5 minutes)
	if im.shouldRefreshCache(categoryStr) {
		if err := im.scanCategoryImages(categoryStr); err != nil {
			return "", fmt.Errorf("failed to scan images for category %d: %w", categoryID, err)
		}
	}

	// Get available images for this category
	images := im.imageCache[categoryStr]
	if len(images) == 0 {
		return "", fmt.Errorf("no images found for category %d", categoryID)
	}

	// Initialize usage stats for this category if not exists
	if im.usageStats[categoryStr] == nil {
		im.usageStats[categoryStr] = make(map[string]int)
	}

	// Find images with minimum usage
	minUsage := -1
	var candidates []string

	for _, imagePath := range images {
		usage := im.usageStats[categoryStr][imagePath]
		if minUsage == -1 || usage < minUsage {
			minUsage = usage
			candidates = []string{imagePath}
		} else if usage == minUsage {
			candidates = append(candidates, imagePath)
		}
	}

	// Select random image from candidates with minimum usage
	selectedImage := candidates[rand.Intn(len(candidates))]

	// Increment usage count
	im.usageStats[categoryStr][selectedImage]++

	return selectedImage, nil
}

// shouldRefreshCache checks if the cache should be refreshed for a category
func (im *ImageManager) shouldRefreshCache(categoryStr string) bool {
	lastScan, exists := im.lastScanTime[categoryStr]
	if !exists {
		return true
	}

	// Refresh every 5 minutes
	return time.Since(lastScan) > 5*time.Minute
}

// scanCategoryImages scans the images directory for the given category
func (im *ImageManager) scanCategoryImages(categoryStr string) error {
	dirPath := filepath.Join(config.ImageDir, categoryStr)

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
		im.imageCache[categoryStr] = []string{}
		im.lastScanTime[categoryStr] = time.Now()
		return nil
	}

	var images []string

	// Walk through the directory to find image files
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Check if it's an image file (common extensions)
			ext := filepath.Ext(path)
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
				// Construct full URL from base path and filename
				fileName := filepath.Base(path)
				fullURL := strings.TrimRight(config.ImagePath, "/") + "/" + categoryStr + "/" + fileName
				images = append(images, fullURL)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory %s: %w", dirPath, err)
	}

	im.imageCache[categoryStr] = images
	im.lastScanTime[categoryStr] = time.Now()

	return nil
}

// GetUsageStats returns usage statistics for debugging/monitoring
func (im *ImageManager) GetUsageStats() map[string]map[string]int {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := make(map[string]map[string]int)
	for category, usage := range im.usageStats {
		stats[category] = make(map[string]int)
		for image, count := range usage {
			stats[category][image] = count
		}
	}

	return stats
}

// ResetUsageStats resets all usage statistics
func (im *ImageManager) ResetUsageStats() {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.usageStats = make(map[string]map[string]int)
}
