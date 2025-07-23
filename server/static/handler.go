package static

import (
	"github.com/dimfeld/httptreemux"
	"journey/filenames"
	"net/http"
	"os"
	"path/filepath"
)

func RegisterHandlers(router *httptreemux.TreeMux) {
	for _, file := range staticFilesList {
		filePath := filepath.Join("/", file)
		router.GET(filePath, Handler)
	}
}

// Handler serves static files from the StaticFilepath directory.
func Handler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var err error

	staticFile, cached := cachedStaticFiles[r.URL.Path]
	if cached {
		w.Header().Set("Content-Type", staticFile.MimeType)
		w.Write(staticFile.Content)
		return
	}

	filePath := filepath.Join(filenames.StaticFilepath, r.URL.Path)
	_, err = os.Stat(filePath)
	if err == nil {
		http.ServeFile(w, r, filePath)
		return
	}

	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}
