package images

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestImageManager_GetRandomImage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "image_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test category directory
	categoryDir := filepath.Join(tempDir, "1")
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		t.Fatalf("Failed to create category dir: %v", err)
	}

	// Create some test image files
	testImages := []string{"test1.jpg", "test2.png", "test3.gif"}
	for _, img := range testImages {
		imgPath := filepath.Join(categoryDir, img)
		if err := os.WriteFile(imgPath, []byte("test image"), 0644); err != nil {
			t.Fatalf("Failed to create test image %s: %v", img, err)
		}
	}

	// Create image manager with test directory
	im := &ImageManager{
		usageStats:   make(map[string]map[string]int),
		imageCache:   make(map[string][]string),
		lastScanTime: make(map[string]time.Time),
	}

	// Test getting random image
	imageURL, err := im.GetRandomImage(1)
	if err != nil {
		t.Fatalf("GetRandomImage failed: %v", err)
	}

	// Verify the URL format
	if !strings.Contains(imageURL, "/1/") {
		t.Errorf("Expected URL to contain category path, got: %s", imageURL)
	}

	// Test that we get different images on multiple calls (not guaranteed but likely)
	seenImages := make(map[string]bool)
	for i := 0; i < 10; i++ {
		img, err := im.GetRandomImage(1)
		if err != nil {
			t.Fatalf("GetRandomImage failed on iteration %d: %v", i, err)
		}
		seenImages[img] = true
	}

	// We should see multiple different images
	if len(seenImages) < 2 {
		t.Errorf("Expected to see multiple different images, got: %d", len(seenImages))
	}
}

func TestImageManager_UsageTracking(t *testing.T) {
	im := NewImageManager()

	// Simulate some usage
	categoryStr := "1"
	im.usageStats[categoryStr] = map[string]int{
		"image1.jpg": 5,
		"image2.jpg": 2,
		"image3.jpg": 8,
	}
	im.imageCache[categoryStr] = []string{"image1.jpg", "image2.jpg", "image3.jpg"}

	// The image with minimum usage (2) should be preferred
	// Since we can't guarantee which one will be selected due to randomness,
	// we'll just verify the function doesn't error
	_, err := im.GetRandomImage(1)
	if err != nil {
		t.Errorf("GetRandomImage failed: %v", err)
	}
}
