package templates

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/kabukky/journey/configuration"
	"github.com/kabukky/journey/database"
	"github.com/kabukky/journey/structure/methods"
)

type URL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

func ShowSitemap(w http.ResponseWriter) error {
	methods.Blog.RLock()
	defer methods.Blog.RUnlock()

	baseURL := configuration.Config.Url
	if configuration.Config.HttpsUsage != "None" {
		baseURL = configuration.Config.HttpsUrl
	}

	urlset := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []URL{},
	}

	urlset.URLs = append(urlset.URLs, URL{
		Loc:        baseURL,
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	posts, err := database.RetrievePostsForIndex(9999, 0)
	if err != nil {
		return err
	}

	for _, post := range posts {
		if post.IsPublished && !post.IsPage {
			urlset.URLs = append(urlset.URLs, URL{
				Loc:        baseURL + "/" + post.Slug + "/",
				LastMod:    post.Date.Format(time.RFC3339),
				ChangeFreq: "weekly",
				Priority:   "0.8",
			})
		}
	}

	for _, post := range posts {
		if post.IsPublished && post.IsPage {
			urlset.URLs = append(urlset.URLs, URL{
				Loc:        baseURL + "/" + post.Slug + "/",
				LastMod:    post.Date.Format(time.RFC3339),
				ChangeFreq: "monthly",
				Priority:   "0.6",
			})
		}
	}

	tags, err := database.RetrieveAllTags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		urlset.URLs = append(urlset.URLs, URL{
			Loc:        baseURL + "/tag/" + tag.Slug + "/",
			ChangeFreq: "weekly",
			Priority:   "0.5",
		})
	}

	users, err := database.RetrieveAllUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		urlset.URLs = append(urlset.URLs, URL{
			Loc:        baseURL + "/author/" + user.Slug + "/",
			ChangeFreq: "weekly",
			Priority:   "0.5",
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	return encoder.Encode(urlset)
}
