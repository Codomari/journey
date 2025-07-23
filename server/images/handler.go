package images

import (
	"bytes"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"journey/compression"
	"journey/filenames"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	imagePath := filepath.Join(filenames.ImagesFilepath, params["filepath"])

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check for resize parameters
	maxWidthStr := r.URL.Query().Get("maxWidth")
	maxHeightStr := r.URL.Query().Get("maxHeight")

	var maxWidth, maxHeight int
	if maxWidthStr != "" {
		maxWidth, _ = strconv.Atoi(maxWidthStr)
	}
	if maxHeightStr != "" {
		maxHeight, _ = strconv.Atoi(maxHeightStr)
	}

	// Apply parameter copying logic
	if maxWidth > 0 && maxHeightStr == "" {
		maxHeight = maxWidth
	}
	if maxHeight > 0 && maxWidthStr == "" {
		maxWidth = maxHeight
	}

	// Determine if we should resize
	shouldResize := (maxHeight > 0 || maxWidth > 0) && compression.IsImageFile(imagePath)

	// Handle image resizing if needed
	if shouldResize {
		if handleImageResize(w, r, imagePath, maxWidth, maxHeight, fileInfo) {
			return
		}
	}

	// Try to serve compressed version with caching if it's an image file
	if compression.IsImageFile(imagePath) {
		compressedData, wasFromCache, err := compression.CompressImageWithCache(imagePath, filenames.ImagesCacheFilepath)
		if err == nil {
			// Generate ETag for compressed content
			etag := fmt.Sprintf(`"%x-%x-compressed"`, fileInfo.ModTime().Unix(), len(compressedData))
			w.Header().Set("ETag", etag)

			// Check If-None-Match header for ETag validation
			if match := r.Header.Get("If-None-Match"); match != "" {
				if match == etag {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			// Set appropriate content type
			contentType := mime.TypeByExtension(filepath.Ext(imagePath))
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}

			// Set cache headers
			w.Header().Set("Cache-Control", "public, max-age=7776000")

			// Add compression info header for debugging
			if wasFromCache {
				w.Header().Set("X-Compression-Cache", "hit")
			} else {
				w.Header().Set("X-Compression-Cache", "miss")
			}

			// Serve compressed content
			w.Write(compressedData)
			return
		}
	}

	// Fallback to original file serving if compression fails
	// Generate ETag based on file modification time and size
	// Format: "modtime-size" (similar to Apache's default ETag format)
	etag := fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
	w.Header().Set("ETag", etag)

	w.Header().Set("Cache-Control", "public, max-age=7776000")

	// Check If-None-Match header for ETag validation
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	http.ServeFile(w, r, imagePath)
	return
}

func handleImageResize(w http.ResponseWriter, r *http.Request, imagePath string, maxWidth, maxHeight int, fileInfo os.FileInfo) bool {
	resizedData, wasFromCache, err := resizeImageWithCache(imagePath, maxWidth, maxHeight)
	if err != nil {
		return false
	}

	// Generate ETag for resized content
	etag := fmt.Sprintf(`"%x-%dx%d-resized"`, fileInfo.ModTime().Unix(), maxWidth, maxHeight)
	w.Header().Set("ETag", etag)

	// Check If-None-Match header for ETag validation
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}

	// Set appropriate content type
	contentType := mime.TypeByExtension(filepath.Ext(imagePath))
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Set cache headers
	w.Header().Set("Cache-Control", "public, max-age=7776000")

	// Add resize info header for debugging
	if wasFromCache {
		w.Header().Set("X-Resize-Cache", "hit")
	} else {
		w.Header().Set("X-Resize-Cache", "miss")
	}

	// Serve resized content
	w.Write(resizedData)
	return true
}

func resizeImageWithCache(imagePath string, maxWidth, maxHeight int) ([]byte, bool, error) {
	// Generate cache filename with the required extension format: $filename.$ext.320
	ext := filepath.Ext(imagePath)
	baseName := strings.TrimSuffix(filepath.Base(imagePath), ext)
	cacheFilename := fmt.Sprintf("%s%s.320", baseName, ext)
	cachePath := filepath.Join(filenames.ImagesCacheFilepath, cacheFilename)

	// Check if cached resized image exists and is newer than original
	if cacheInfo, err := os.Stat(cachePath); err == nil {
		if origInfo, err := os.Stat(imagePath); err == nil {
			if cacheInfo.ModTime().After(origInfo.ModTime()) || cacheInfo.ModTime().Equal(origInfo.ModTime()) {
				// Serve from cache
				data, err := os.ReadFile(cachePath)
				if err == nil {
					return data, true, nil
				}
			}
		}
	}

	// Open and decode the original image
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	var img image.Image
	if strings.ToLower(ext) == ".png" {
		img, err = png.Decode(file)
	} else if strings.ToLower(ext) == ".jpg" || strings.ToLower(ext) == ".jpeg" {
		img, err = jpeg.Decode(file)
	} else {
		return nil, false, fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return nil, false, err
	}

	// Resize the image maintaining aspect ratio
	var resizedImg image.Image
	if maxWidth > 0 && maxHeight > 0 {
		resizedImg = resize.Thumbnail(uint(maxWidth), uint(maxHeight), img, resize.Lanczos3)
	} else if maxWidth > 0 {
		resizedImg = resize.Resize(uint(maxWidth), 0, img, resize.Lanczos3)
	} else if maxHeight > 0 {
		resizedImg = resize.Resize(0, uint(maxHeight), img, resize.Lanczos3)
	} else {
		return nil, false, fmt.Errorf("no resize dimensions specified")
	}

	// Encode the resized image to bytes
	var buf bytes.Buffer
	if strings.ToLower(ext) == ".png" {
		err = png.Encode(&buf, resizedImg)
	} else {
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return nil, false, err
	}

	resizedData := buf.Bytes()

	// Save to cache
	err = os.WriteFile(cachePath, resizedData, 0644)
	if err != nil {
		// Don't fail if we can't write to cache, just return the resized data
		return resizedData, false, nil
	}

	return resizedData, false, nil
}
