package thumb

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/webp"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)
}

var (
	// Size is max width or height of the thumbnail image
	Size uint = 256
	// Quality is the thumbnail jpeg quality
	Quality int = 100
)

// Get reads an image from an io.Reader and returns the thumbnail in bytes
func Get(img io.Reader, label string) ([]byte, error) {
	src, _, err := image.Decode(img)
	if err != nil {
		return nil, err
	}

	dst := resize.Thumbnail(Size, Size, src, resize.NearestNeighbor)

	if len(label) > 0 {
		dst = toDrawImage(dst)

		b := dst.Bounds()
		width := b.Dx() - 16
		height := b.Dy() - 16
		maxLen := width / 7

		if width > 24 && height > 24 {
			if len(label) > maxLen {
				label = label[:maxLen] + ".."
			}
			addLabel(dst, 8, height+8, color.Black, label)
			addLabel(dst, 6, height+6, color.White, label)
		}
	}

	var result bytes.Buffer
	err = jpeg.Encode(&result, dst, &jpeg.Options{Quality: Quality})
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// GetFromURL downloads the image at the given URL and returns the thumbnail in bytes
func GetFromURL(ctx context.Context, url, label string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "image/*")

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-type")
	t, _, err := mime.ParseMediaType(contentType)
	if !strings.HasPrefix(t, "image/") {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	return Get(resp.Body, label)
}

func toDrawImage(src image.Image) draw.Image {
	dst, ok := src.(draw.Image)
	if ok {
		return dst
	}

	b := src.Bounds()
	dst = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func addLabel(img image.Image, x, y int, col color.Color, label string) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	d := &font.Drawer{
		Dst:  img.(draw.Image),
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}
