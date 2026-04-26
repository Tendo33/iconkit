package processor

import (
	"image"
	"image/color"
	"os"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// RasterizeSVG rasterizes an SVG file to an image.Image at a fixed 512x512 size.
// The caller is responsible for further resizing.
func RasterizeSVG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	icon, err := oksvg.ReadIconStream(f)
	if err != nil {
		return nil, err
	}

	const renderSize = 512
	icon.SetTarget(0, 0, float64(renderSize), float64(renderSize))

	rgba := image.NewRGBA(image.Rect(0, 0, renderSize, renderSize))

	// Fill with transparent background
	for y := 0; y < renderSize; y++ {
		for x := 0; x < renderSize; x++ {
			rgba.Set(x, y, color.Transparent)
		}
	}

	scanner := rasterx.NewScannerGV(renderSize, renderSize, rgba, rgba.Bounds())
	rasterizer := rasterx.NewDasher(renderSize, renderSize, scanner)
	icon.Draw(rasterizer, 1.0)

	return rgba, nil
}
