//go:build headless

// Package video provides stubs for headless/server builds that lack FFmpeg/x264.
package video

import (
	"fmt"
	"image"
)

// Encoder stub — not usable in headless mode.
type Encoder struct {
	Width  int
	Height int
}

func NewEncoder(width, height int) (*Encoder, error) {
	return nil, fmt.Errorf("video codec not available in headless build")
}
func (e *Encoder) Encode(_ image.Image) ([]byte, error) { return nil, fmt.Errorf("headless") }
func (e *Encoder) Close()                               {}

// Decoder stub.
type Decoder struct{}

func NewDecoder() (*Decoder, error)                     { return nil, fmt.Errorf("headless") }
func (d *Decoder) Decode(_ []byte) (image.Image, error) { return nil, fmt.Errorf("headless") }
func (d *Decoder) Close()                               {}
