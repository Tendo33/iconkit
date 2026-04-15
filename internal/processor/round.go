package processor

import (
	"image"
	"image/color"
	"math"
)

// RoundCorners applies rounded corners with the given radius (in pixels) to the image.
// Pixels outside the rounded rectangle become fully transparent.
func RoundCorners(img image.Image, radius int) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if radius <= 0 {
		return toNRGBA(img)
	}

	// Clamp radius to half of the smallest dimension
	maxR := w / 2
	if h/2 < maxR {
		maxR = h / 2
	}
	if radius > maxR {
		radius = maxR
	}

	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	r := float64(radius)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if inRoundedRect(x, y, w, h, r) {
				dst.Set(x, y, img.At(x+bounds.Min.X, y+bounds.Min.Y))
			} else {
				dst.Set(x, y, color.Transparent)
			}
		}
	}
	return dst
}

// inRoundedRect checks if point (x,y) is inside a rounded rectangle of size w×h.
func inRoundedRect(x, y, w, h int, r float64) bool {
	fx, fy := float64(x)+0.5, float64(y)+0.5
	fw, fh := float64(w), float64(h)

	// Only corners need the circle check
	var cx, cy float64
	switch {
	case fx < r && fy < r:
		cx, cy = r, r
	case fx > fw-r && fy < r:
		cx, cy = fw-r, r
	case fx < r && fy > fh-r:
		cx, cy = r, fh-r
	case fx > fw-r && fy > fh-r:
		cx, cy = fw-r, fh-r
	default:
		return true
	}

	dx := fx - cx
	dy := fy - cy
	return dx*dx+dy*dy <= r*r+0.5 // small epsilon for anti-alias edge
}

// ScaleRadius converts a radius value relative to the original image size
// to the appropriate radius for a target size.
func ScaleRadius(originalRadius, originalSize, targetSize int) int {
	if originalSize == 0 {
		return 0
	}
	scaled := float64(originalRadius) * float64(targetSize) / float64(originalSize)
	return int(math.Round(scaled))
}

func toNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, img.At(x, y))
		}
	}
	return dst
}
