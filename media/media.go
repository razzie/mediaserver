package media

import (
	"context"
	"fmt"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"

	"github.com/razzie/mediaserver/siteinfo"
	"github.com/razzie/mediaserver/thumb"
)

// Media contains basic details about a website and a thumbnail
type Media struct {
	SiteInfo  *siteinfo.SiteInfo `json:"siteinfo"`
	Thumbnail *thumb.Thumbnail   `json:"thumbnail"`
}

// GetFromURL tries to get media data from an URL
func GetFromURL(ctx context.Context, url string) (*Media, error) {
	req, err := http.NewRequest("GET", "http://"+url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	m := &Media{}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		m.Thumbnail, err = thumb.Get(resp.Body, "")
		return m, err
	}

	m.SiteInfo, err = siteinfo.Get(resp.Body)
	if err != nil {
		return m, err
	}

	if len(m.SiteInfo.Images) == 0 {
		return m, fmt.Errorf("no thumbnail available")
	}

	m.SiteInfo.ResolveImageURLs("http://" + url)

	for _, img := range m.SiteInfo.Images {
		m.Thumbnail, err = thumb.GetFromURL(ctx, img, m.SiteInfo.Title)
		if err == nil {
			return m, nil
		}
	}

	return m, err
}

func (m Media) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.Thumbnail == nil {
		http.Error(w, "no thumbnail available", http.StatusNotFound)
		return
	}

	if m.SiteInfo != nil {
		crc := crc32.ChecksumIEEE([]byte(m.SiteInfo.URL))
		filename := strconv.FormatUint(uint64(crc), 36) + "." + m.Thumbnail.MIME[6:]
		w.Header().Set("Content-Disposition", "filename="+filename)
	}

	m.Thumbnail.ServeHTTP(w, r)
}
