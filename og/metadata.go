package og

import (
	"context"
	"io"
	"net/http"

	"github.com/otiai10/opengraph"
)

// Metadata holds the most typical OpenGraph metadata
type Metadata struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image"`
}

func newMetadata(og *opengraph.OpenGraph) *Metadata {
	var image string
	if len(og.Image) > 0 {
		image = og.Image[0].URL
	}

	return &Metadata{
		Type:        og.Type,
		URL:         og.URL.String(),
		Title:       og.Title,
		Description: og.Description,
		ImageURL:    image,
	}
}

// Get returns metadata from an io.Reader that contains HTML
func Get(url string, html io.Reader) (*Metadata, error) {
	og := opengraph.New(url)
	err := og.Parse(html)
	if err != nil {
		return nil, err
	}

	return newMetadata(og), nil
}

// GetFromURL returns metadata from an URL
func GetFromURL(ctx context.Context, url, useragent string) (*Metadata, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("accept", "text/html")

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return Get(url, resp.Body)
}
