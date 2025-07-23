package compression

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestPNG creates a simple test PNG image
func createTestPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with a simple pattern
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
			} else {
				img.Set(x, y, color.RGBA{0, 255, 0, 255}) // Green
			}
		}
	}
	
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestCompressImageLossless(t *testing.T) {
	testPNG := createTestPNG()
	
	// Test PNG compression
	compressed, wasCompressed, err := CompressImageLossless(testPNG, "test.png")
	if err != nil {
		t.Fatalf("PNG compression failed: %v", err)
	}
	
	// Compressed data should be valid
	if len(compressed) == 0 {
		t.Error("Compressed data is empty")
	}
	
	// For unknown formats, should return original data
	unknownData := []byte("not an image")
	result, wasCompressed, err := CompressImageLossless(unknownData, "test.txt")
	if err != nil {
		t.Fatalf("Unknown format handling failed: %v", err)
	}
	
	if wasCompressed {
		t.Error("Unknown format should not be compressed")
	}
	
	if !bytes.Equal(result, unknownData) {
		t.Error("Unknown format data was modified")
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.png", true},
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.gif", true},
		{"test.svg", true},
		{"test.txt", false},
		{"test.pdf", false},
		{"test", false},
	}
	
	for _, test := range tests {
		result := IsImageFile(test.filename)
		if result != test.expected {
			t.Errorf("IsImageFile(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestGetImageFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.png", "png"},
		{"test.jpg", "jpeg"},
		{"test.jpeg", "jpeg"},
		{"test.gif", "gif"},
		{"test.svg", "svg"},
		{"test.txt", "unknown"},
	}
	
	for _, test := range tests {
		result := GetImageFormat(test.filename)
		if result != test.expected {
			t.Errorf("GetImageFormat(%s) = %s, expected %s", test.filename, result, test.expected)
		}
	}
}

func TestCompressImageWithCache(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "compression_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test image file
	testPNG := createTestPNG()
	testFile := filepath.Join(tempDir, "test.png")
	err = os.WriteFile(testFile, testPNG, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	cacheDir := filepath.Join(tempDir, "cache")
	
	// First call should compress and cache
	compressed1, fromCache1, err := CompressImageWithCache(testFile, cacheDir)
	if err != nil {
		t.Fatalf("First compression failed: %v", err)
	}
	
	if fromCache1 {
		t.Error("First call should not be from cache")
	}
	
	if len(compressed1) == 0 {
		t.Error("Compressed data is empty")
	}
	
	// Second call should use cache
	compressed2, fromCache2, err := CompressImageWithCache(testFile, cacheDir)
	if err != nil {
		t.Fatalf("Second compression failed: %v", err)
	}
	
	if !fromCache2 {
		t.Error("Second call should be from cache")
	}
	
	if !bytes.Equal(compressed1, compressed2) {
		t.Error("Cached and fresh compressed data don't match")
	}
}

func TestCleanupCache(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create some test cache files
	oldFile := filepath.Join(tempDir, "old.png.compressed")
	newFile := filepath.Join(tempDir, "new.png.compressed")
	regularFile := filepath.Join(tempDir, "regular.txt")
	
	// Create files with different ages
	err = os.WriteFile(oldFile, []byte("old"), 0644)
	if err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	
	err = os.WriteFile(newFile, []byte("new"), 0644)
	if err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}
	
	err = os.WriteFile(regularFile, []byte("regular"), 0644)
	if err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}
	
	// Make old file actually old
	oldTime := time.Now().Add(-8 * 24 * time.Hour) // 8 days old
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to change old file time: %v", err)
	}
	
	// Run cleanup (remove files older than 7 days)
	err = CleanupCache(tempDir, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	
	// Old compressed file should be gone
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old compressed file should have been removed")
	}
	
	// New compressed file should still exist
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New compressed file should still exist")
	}
	
	// Regular file should still exist
	if _, err := os.Stat(regularFile); os.IsNotExist(err) {
		t.Error("Regular file should still exist")
	}
}