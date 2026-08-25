//go:build headless

package host

import "image"

func captureRaw(maxWidth, maxHeight int) (image.Image, int, int, error) {
	return nil, 0, 0, nil
}

func screenSize() (int, int) { return 1920, 1080 }
