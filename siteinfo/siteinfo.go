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
	Type        string `json:"type"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image"`
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
		s.ImageURL = content
	}
}

// ResolveImageURL resolves a potentially relative image URL using the given hostname
func (s *SiteInfo) ResolveImageURL(hostname string) error {
	if strings.Contains(s.ImageURL, "://") {
		return nil
	}

	imgURL, err := url.Parse(s.ImageURL)
	if err != nil {
		return err
	}

	hostURL, err := url.Parse(hostname)
	if err != nil {
		return err
	}

	s.ImageURL = hostURL.ResolveReference(imgURL).String()
	return nil
}

func (s *SiteInfo) isReady() bool {
	return len(s.Title) > 0 && len(s.ImageURL) > 0
}

// Get returns SiteInfo from an io.Reader that contains HTML
func Get(buffer io.Reader) (*SiteInfo, error) {
	s := &SiteInfo{}
	z := html.NewTokenizer(buffer)
	base := ""
	inBody := false
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
			case atom.Body:
				inBody = true

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
					if len(s.ImageURL) == 0 && !hasParent(atom.A) {
						url := m["src"]
						if strings.Contains(url, "://") {
							s.ImageURL = url
						} else {
							s.ImageURL = path.Join(base, url)
						}
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
			if !inBody && len(s.Title) == 0 && hasParent(atom.Title) {
				s.Title = string(z.Text())
			}
		}

		if inBody && s.isReady() {
			break
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
