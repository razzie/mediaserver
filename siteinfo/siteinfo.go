package siteinfo

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// SiteInfo holds the most typical details about a website (if found)
type SiteInfo struct {
	Type        string   `json:"type"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
}

// ProcessMeta updates the SiteInfo based on OpenGraph property name and content
func (s *SiteInfo) ProcessMeta(property, content string) {
	if len(content) == 0 {
		return
	}

	switch property {
	case "og:description":
		s.Description = content
	case "og:type":
		s.Type = content
	case "og:title":
		s.Title = content
	case "og:url":
		s.URL = content
	case "og:image", "og:image:url":
		s.Images = append(s.Images, content)
	}
}

func resolveImageURL(img *string, hostname string) error {
	if strings.Contains(*img, "://") {
		return nil
	}

	if strings.HasPrefix(*img, "//") {
		*img = "http:" + *img
		return nil
	}

	imgURL, err := url.Parse(*img)
	if err != nil {
		return err
	}

	hostURL, err := url.Parse(hostname)
	if err != nil {
		return err
	}

	*img = hostURL.ResolveReference(imgURL).String()
	return nil
}

// ResolveImageURLs resolves a potentially relative image URLs using the given hostname
func (s *SiteInfo) ResolveImageURLs(hostname string) {
	for i := 0; i < len(s.Images); i++ {
		img := &s.Images[i]
		resolveImageURL(img, hostname)
	}
}

// Get returns SiteInfo from an io.Reader that contains HTML
func Get(buffer io.Reader) (*SiteInfo, error) {
	s := &SiteInfo{}
	z := html.NewTokenizer(buffer)
	base := ""
	parents := make([]atom.Atom, 0, 10)
	hasParent := func(p atom.Atom) bool {
		for _, a := range parents {
			if a == p {
				return true
			}
		}
		return false
	}

tokenize:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				break tokenize
			}
			return s, z.Err()

		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			a := atom.Lookup(name)

			switch a {
			case atom.Meta, atom.Img, atom.Base:
				m := make(map[string]string)
				var key, val []byte
				for hasAttr {
					key, val, hasAttr = z.TagAttr()
					m[atom.String(key)] = string(val)
				}

				switch a {
				case atom.Meta:
					if len(s.Description) == 0 && m["name"] == "description" {
						m["property"] = "og:description"
					}
					s.ProcessMeta(m["property"], m["content"])

				case atom.Img:
					if !hasParent(atom.A) {
						url := m["src"]
						if !strings.Contains(url, "://") && !strings.HasPrefix(url, "//") {
							url = path.Join(base, url)
						}
						s.Images = append(s.Images, url)
					}

				case atom.Base:
					base = m["href"]
				}
			}

			if tt == html.StartTagToken {
				parents = append(parents, a)
			}

		case html.EndTagToken:
			name, _ := z.TagName()
			a := atom.Lookup(name)
			for i := len(parents) - 1; i >= 0; i-- {
				if parents[i] == a {
					parents = parents[:i]
					break
				}
			}

		case html.TextToken:
			if len(s.Title) == 0 && hasParent(atom.Title) {
				s.Title = string(z.Text())
			}
		}
	}

	return s, nil
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

	s, err := Get(resp.Body)
	if s != nil && len(s.URL) == 0 {
		s.URL = url
	}

	return s, err
}
