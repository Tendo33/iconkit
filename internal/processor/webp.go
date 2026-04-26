package processor

import (
	"fmt"
	"image"
	"os"

	"golang.org/x/image/webp"
)

// DecodeWebP decodes a WebP file into an image.Image using the pure-Go decoder.
func DecodeWebP(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open WebP file: %w", err)
	}
	defer f.Close()

	img, err := webp.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WebP: %w", err)
	}
	return img, nil
}

// EncodeWebP encodes an image.Image to a WebP file.
// WebP encoding requires CGO, which is not available in all build environments.
// In CGO-disabled builds, this returns an error with guidance to use --format png.
func EncodeWebP(img image.Image, path string, quality float32) error {
	return fmt.Errorf(
		"WebP encoding requires a CGO-enabled build and is not available in this binary.\n" +
			"Use --format png for the output format, or build iconkit with CGO_ENABLED=1.",
	)
}
