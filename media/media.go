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
	SiteInfo      *siteinfo.SiteInfo `json:"siteinfo"`
	Thumbnail     []byte             `json:"thumbnail"`
	ThumbnailMIME string             `json:"thumbnail_mime"`
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
		m.Thumbnail, m.ThumbnailMIME, err = thumb.Get(resp.Body, "")
		return m, err
	}

	m.SiteInfo, err = siteinfo.Get(resp.Body)
	if err != nil {
		return m, err
	}

	if len(m.SiteInfo.ImageURL) == 0 {
		return m, fmt.Errorf("no thumbnail available")
	}

	if err = m.SiteInfo.ResolveImageURL("http://" + url); err != nil {
		return m, err
	}

	m.Thumbnail, m.ThumbnailMIME, err = thumb.GetFromURL(ctx, m.SiteInfo.ImageURL, m.SiteInfo.Title)
	return m, err
}

func (m Media) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(m.Thumbnail) == 0 {
		http.Error(w, "no thumbnail available", http.StatusNotFound)
		return
	}

	if m.SiteInfo != nil {
		crc := crc32.ChecksumIEEE([]byte(m.SiteInfo.URL))
		filename := strconv.FormatUint(uint64(crc), 36) + "." + m.ThumbnailMIME[6:]
		w.Header().Set("Content-Disposition", "filename="+filename)
	}

	w.Header().Set("Content-Type", m.ThumbnailMIME)
	w.Header().Set("Content-Length", strconv.Itoa(len(m.Thumbnail)))
	w.Write(m.Thumbnail)
}
