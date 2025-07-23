package compression

import (
	"log"
	"time"
)

// StartCacheCleanup starts a background goroutine to periodically clean up old cache files
func StartCacheCleanup(cacheDir string, cleanupInterval time.Duration, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				err := CleanupCache(cacheDir, maxAge)
				if err != nil {
					log.Printf("Error cleaning up image cache: %v", err)
				} else {
					log.Printf("Image cache cleanup completed successfully")
				}
			}
		}
	}()
}