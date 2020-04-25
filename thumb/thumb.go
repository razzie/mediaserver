package thumb

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"

	"github.com/nfnt/resize"
	"golang.org/x/image/webp"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)

}

// Get reads an image from an io.Reader and returns the thumbnail in bytes
func Get(img io.Reader) ([]byte, error) {
	src, _, err := image.Decode(img)
	if err != nil {
		return nil, err
	}

	dst := resize.Thumbnail(256, 256, src, resize.NearestNeighbor)
	var result bytes.Buffer

	err = jpeg.Encode(&result, dst, &jpeg.Options{Quality: 10})
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// GetFromURL downloads the image at the given URL and returns the thumbnail in bytes
func GetFromURL(ctx context.Context, url, useragent string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("accept", "image/*")

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-type")
	t, _, err := mime.ParseMediaType(contentType)
	if t != "image" {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	return Get(resp.Body)
}