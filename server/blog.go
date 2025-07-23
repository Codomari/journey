package server

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"journey/database"
	"journey/filenames"
	"journey/structure/methods"
	"journey/templates"

	"github.com/dimfeld/httptreemux"
)

func indexHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	number := params["number"]
	if number == "" {
		// Render index template (first page)
		err := templates.ShowIndexTemplate(w, r, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	page, err := strconv.Atoi(number)
	if err != nil || page <= 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Render index template
	err = templates.ShowIndexTemplate(w, r, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func authorHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]
	function := params["function"]
	number := params["number"]
	if function == "" {
		// Render author template (first page)
		err := templates.ShowAuthorTemplate(w, r, slug, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	} else if function == "rss" {
		// Render author rss feed
		err := templates.ShowAuthorRss(w, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	page, err := strconv.Atoi(number)
	if err != nil || page <= 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Render author template
	err = templates.ShowAuthorTemplate(w, r, slug, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func tagHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]
	function := params["function"]
	number := params["number"]
	if function == "" {
		// Render tag template (first page)
		err := templates.ShowTagTemplate(w, r, slug, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	} else if function == "rss" {
		// Render tag rss feed
		err := templates.ShowTagRss(w, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	page, err := strconv.Atoi(number)
	if err != nil || page <= 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Render tag template
	err = templates.ShowTagTemplate(w, r, slug, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func postHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else if slug == "rss" {
		// Render index rss feed
		err := templates.ShowIndexRss(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Render post template
	err := templates.ShowPostTemplate(w, r, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func postEditHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	slug := params["slug"]

	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Redirect to edit
	post, err := database.RetrievePostBySlug(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("/admin/#/edit/%d", post.Id)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func assetsHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	methods.Blog.RLock()
	defer methods.Blog.RUnlock()
	http.ServeFile(w, r, filepath.Join(filenames.ThemesFilepath, methods.Blog.ActiveTheme, "assets", params["filepath"]))
	return
}

func imagesHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	imagePath := filepath.Join(filenames.ImagesFilepath, params["filepath"])

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

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

func publicHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	http.ServeFile(w, r, filepath.Join(filenames.PublicFilepath, params["filepath"]))
	return
}

func sitemapHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	err := templates.ShowSitemap(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// staticHandler serves static files from the StaticFilepath directory.
func staticHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
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

type StaticFile struct {
	Name     string
	Path     string
	Ext      string
	Content  []byte
	MimeType string
}

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

func InitializeBlog(router *httptreemux.TreeMux) {
	// For index
	router.GET("/", indexHandler)
	router.GET("/:slug/edit", postEditHandler)
	router.GET("/:slug/", postHandler)
	router.GET("/page/:number/", indexHandler)
	// For author
	router.GET("/author/:slug/", authorHandler)
	router.GET("/author/:slug/:function/", authorHandler)
	router.GET("/author/:slug/:function/:number/", authorHandler)
	// For tag
	router.GET("/tag/:slug/", tagHandler)
	router.GET("/tag/:slug/:function/", tagHandler)
	router.GET("/tag/:slug/:function/:number/", tagHandler)
	// For serving asset files
	router.GET("/assets/*filepath", assetsHandler)
	router.GET("/images/*filepath", imagesHandler)
	router.GET("/content/images/*filepath", imagesHandler) // This is here to keep compatibility with Ghost
	router.GET("/public/*filepath", publicHandler)
	// For sitemap
	router.GET("/sitemap.xml", sitemapHandler)
	// For static files
	for _, file := range staticFilesList {
		filePath := filepath.Join("/", file)
		router.GET(filePath, staticHandler)
	}
}
