//go:build !headless

package video

import (
    "bytes"
    "fmt"
    "image"
    "image/jpeg"
)

// Encoder encodes raw screen images to MJPEG (JPEG frames).
type Encoder struct {
    Width  int
    Height int
}

func NewEncoder(width, height int) (*Encoder, error) {
    return &Encoder{
        Width:  width,
        Height: height,
    }, nil
}

func (e *Encoder) Encode(img image.Image) ([]byte, error) {
    var buf bytes.Buffer
    if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 70}); err != nil {
        return nil, fmt.Errorf("jpeg encode: %w", err)
    }
    return buf.Bytes(), nil
}

func (e *Encoder) Close() {}

// Decoder decodes MJPEG (JPEG frames) back to image.Image.
type Decoder struct{}

func NewDecoder() (*Decoder, error) {
    return &Decoder{}, nil
}

func (d *Decoder) Decode(data []byte) (image.Image, error) {
    if len(data) == 0 {
        return nil, nil
    }
    img, err := jpeg.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, fmt.Errorf("jpeg decode: %w", err)
    }
    return img, nil
}

func (d *Decoder) Close() {}
