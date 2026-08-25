//go:build windows
package video

import (
    "fmt"
    "image"
)

// Encoder stub for Windows (no FFmpeg)
type Encoder struct {
    Width  int
    Height int
}

func NewEncoder(width, height int) (*Encoder, error) {
    return nil, fmt.Errorf("video codec not available on Windows (FFmpeg disabled)")
}

func (e *Encoder) Encode(_ image.Image) ([]byte, error) {
    return nil, fmt.Errorf("windows stub: no FFmpeg")
}

func (e *Encoder) Close() {}

// Decoder stub for Windows
type Decoder struct{}

func NewDecoder() (*Decoder, error) {
    return nil, fmt.Errorf("windows stub: no FFmpeg")
}

func (d *Decoder) Decode(_ []byte) (image.Image, error) {
    return nil, fmt.Errorf("windows stub: no FFmpeg")
}

func (d *Decoder) Close() {}
