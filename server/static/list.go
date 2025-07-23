package static

import (
	"fmt"
	"journey/filenames"
	"mime"
	"os"
	"path/filepath"
)

var staticFilesList = []string{
	"favicon.ico",
	"robots.txt",
	"android-chrome-192x192.png",
	"android-chrome-512x512.png",
	"apple-touch-icon.png",
	"favicon-16x16.png",
	"favicon-32x32.png",
}

// cachedStaticFiles holds the contents of static files in memory for quick access.
var cachedStaticFiles = map[string]StaticFile{}

func init() {
	for _, file := range staticFilesList {
		filePath := filepath.Join(filenames.StaticFilepath, file)
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading static file %s: %v\n", file, err)
			continue
		}

		ext := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(ext)

		cachedStaticFiles[file] = StaticFile{
			Name:     file,
			Path:     filePath,
			Content:  data,
			MimeType: mimeType,
		}
	}
}
