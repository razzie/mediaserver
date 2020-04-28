package siteinfo

import (
	"context"
	"io"
	"net/http"

	"github.com/dyatlov/go-opengraph/opengraph"
)

// SiteInfo holds the most typical details about a website (if found)
type SiteInfo struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image"`
}

func newSiteInfo(og *opengraph.OpenGraph) *SiteInfo {
	var image string
	if len(og.Images) > 0 {
		image = og.Images[0].URL
	}

	return &SiteInfo{
		Type:        og.Type,
		URL:         og.URL,
		Title:       og.Title,
		Description: og.Description,
		ImageURL:    image,
	}
}

// Get returns SiteInfo from an io.Reader that contains HTML
func Get(html io.Reader) (*SiteInfo, error) {
	og := opengraph.NewOpenGraph()
	err := og.ProcessHTML(html)
	if err != nil {
		return nil, err
	}

	return newSiteInfo(og), nil
}

// GetFromURL returns SiteInfo from an URL
func GetFromURL(ctx context.Context, url string) (*SiteInfo, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "text/html")

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return Get(resp.Body)
}
