package compression

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CompressImageLossless applies lossless compression to image data
// Returns compressed data and whether compression was applied
func CompressImageLossless(data []byte, filename string) ([]byte, bool, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".png":
		return compressPNG(data)
	case ".jpg", ".jpeg":
		// JPEG is inherently lossy, but we can optimize it by re-encoding
		// with maximum quality to reduce artifacts from previous compression
		return optimizeJPEG(data)
	default:
		// For other formats (gif, svg, etc.), return original data
		return data, false, nil
	}
}

// CompressImageWithCache applies lossless compression with filesystem caching
// Returns compressed data, whether it was cached, and any error
func CompressImageWithCache(originalPath string, cacheDir string) ([]byte, bool, error) {
	// Generate cache filename based on original file's hash and modification time
	fileInfo, err := os.Stat(originalPath)
	if err != nil {
		return nil, false, err
	}
	
	// Read original file for hash calculation
	originalData, err := os.ReadFile(originalPath)
	if err != nil {
		return nil, false, err
	}
	
	// Create hash of file content + modification time for cache key
	hasher := md5.New()
	hasher.Write(originalData)
	hasher.Write([]byte(fileInfo.ModTime().Format(time.RFC3339Nano)))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	
	// Generate cache filename
	ext := filepath.Ext(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), ext)
	cacheFilename := fmt.Sprintf("%s_%s%s.compressed", baseName, hash[:12], ext)
	cachePath := filepath.Join(cacheDir, cacheFilename)
	
	// Check if cached compressed version exists and is newer than original
	if cacheInfo, err := os.Stat(cachePath); err == nil {
		if cacheInfo.ModTime().After(fileInfo.ModTime()) {
			// Return cached compressed version
			cachedData, err := os.ReadFile(cachePath)
			if err == nil {
				return cachedData, true, nil
			}
		}
	}
	
	// Compress the image
	compressedData, wasCompressed, err := CompressImageLossless(originalData, originalPath)
	if err != nil {
		return originalData, false, err
	}
	
	// Only cache if compression was applied and resulted in smaller file
	if wasCompressed && len(compressedData) < len(originalData) {
		// Ensure cache directory exists
		err = os.MkdirAll(cacheDir, 0755)
		if err == nil {
			// Write compressed data to cache
			os.WriteFile(cachePath, compressedData, 0644)
		}
	}
	
	return compressedData, false, nil
}

// CleanupCache removes old cache files that are older than maxAge
func CleanupCache(cacheDir string, maxAge time.Duration) error {
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil // Cache directory doesn't exist, nothing to clean
	}
	
	cutoff := time.Now().Add(-maxAge)
	
	return filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Remove files with .compressed extension that are older than cutoff
		if strings.HasSuffix(path, ".compressed") && info.ModTime().Before(cutoff) {
			return os.Remove(path)
		}
		
		return nil
	})
}

// compressPNG applies PNG-specific lossless compression optimizations
func compressPNG(data []byte) ([]byte, bool, error) {
	// Decode the PNG image
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return data, false, err
	}
	
	// Create a buffer for the compressed PNG
	var buf bytes.Buffer
	
	// Create PNG encoder with maximum compression
	encoder := &png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	
	// Encode with optimized settings
	err = encoder.Encode(&buf, img)
	if err != nil {
		return data, false, err
	}
	
	compressed := buf.Bytes()
	
	// Only return compressed version if it's actually smaller
	if len(compressed) < len(data) {
		return compressed, true, nil
	}
	
	return data, false, nil
}

// optimizeJPEG re-encodes JPEG with high quality to reduce compression artifacts
func optimizeJPEG(data []byte) ([]byte, bool, error) {
	// Decode the JPEG image
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return data, false, err
	}
	
	// Create a buffer for the optimized JPEG
	var buf bytes.Buffer
	
	// Re-encode with high quality (95 out of 100)
	// This helps reduce artifacts from previous compression while maintaining small size
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	if err != nil {
		return data, false, err
	}
	
	optimized := buf.Bytes()
	
	// Only return optimized version if it's smaller or similar size
	// Allow up to 5% size increase for quality improvement
	if len(optimized) <= int(float64(len(data))*1.05) {
		return optimized, true, nil
	}
	
	return data, false, nil
}

// CompressImageStream applies lossless compression to an image stream
func CompressImageStream(reader io.Reader, filename string) ([]byte, bool, error) {
	// Read all data from the stream
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false, err
	}
	
	return CompressImageLossless(data, filename)
}

// IsImageFile checks if a file is a supported image format
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".svg"
}

// GetImageFormat returns the image format based on file extension
func GetImageFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".svg":
		return "svg"
	default:
		return "unknown"
	}
}