package processor

import (
	"image"
	"image/color"
	"math"
)

// RoundCorners applies rounded corners with the given radius (in pixels) to the image.
// Pixels outside the rounded rectangle become fully transparent.
// Edge pixels are anti-aliased using 4x4 supersampling.
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
	// Pixels further than this distance from a corner arc center are unambiguously inside or outside.
	innerThresh := r - math.Sqrt2
	outerThresh := r + math.Sqrt2

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cx, cy, inCorner := cornerCenter(x, y, w, h, r)
			if !inCorner {
				// Not near any corner — keep pixel as-is
				dst.Set(x, y, img.At(x+bounds.Min.X, y+bounds.Min.Y))
				continue
			}

			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= innerThresh {
				// Fully inside the arc
				dst.Set(x, y, img.At(x+bounds.Min.X, y+bounds.Min.Y))
			} else if dist >= outerThresh {
				// Fully outside — transparent
				dst.Set(x, y, color.Transparent)
			} else {
				// On the boundary — 4x4 supersampling
				coverage := supersampleCoverage(x, y, cx, cy, r)
				src := img.At(x+bounds.Min.X, y+bounds.Min.Y)
				dst.Set(x, y, applyAlphaCoverage(src, coverage))
			}
		}
	}
	return dst
}

// cornerCenter returns the center of the corner arc for pixel (x,y) if it is
// within the corner region, plus a boolean indicating whether it is in a corner.
func cornerCenter(x, y, w, h int, r float64) (cx, cy float64, inCorner bool) {
	fx, fy := float64(x)+0.5, float64(y)+0.5
	fw, fh := float64(w), float64(h)

	switch {
	case fx < r && fy < r:
		return r, r, true
	case fx > fw-r && fy < r:
		return fw - r, r, true
	case fx < r && fy > fh-r:
		return r, fh - r, true
	case fx > fw-r && fy > fh-r:
		return fw - r, fh - r, true
	default:
		return 0, 0, false
	}
}

// supersampleCoverage returns a value in [0,1] representing how much of the
// pixel at (x,y) lies within the corner arc centered at (cx,cy) with radius r.
// Uses a 4x4 uniform grid of sub-pixel samples.
func supersampleCoverage(x, y int, cx, cy, r float64) float64 {
	const grid = 4
	const step = 1.0 / grid
	inside := 0
	r2 := r * r
	for sy := 0; sy < grid; sy++ {
		for sx := 0; sx < grid; sx++ {
			spx := float64(x) + (float64(sx)+0.5)*step
			spy := float64(y) + (float64(sy)+0.5)*step
			ddx := spx - cx
			ddy := spy - cy
			if ddx*ddx+ddy*ddy <= r2 {
				inside++
			}
		}
	}
	return float64(inside) / float64(grid*grid)
}

// applyAlphaCoverage scales the alpha of src by the given coverage [0,1].
func applyAlphaCoverage(src color.Color, coverage float64) color.Color {
	r, g, b, a := src.RGBA()
	newA := uint8(float64(a>>8) * coverage)
	if a == 0 {
		return color.Transparent
	}
	// Pre-multiplied → straight: src.RGBA() returns pre-multiplied 16-bit values
	var nr, ng, nb uint8
	if a > 0 {
		nr = uint8((r * 0xff / a) & 0xff)
		ng = uint8((g * 0xff / a) & 0xff)
		nb = uint8((b * 0xff / a) & 0xff)
	}
	return color.NRGBA{R: nr, G: ng, B: nb, A: newA}
}

// inRoundedRect checks if point (x,y) is inside a rounded rectangle of size w×h.
// Kept for use in tests.
func inRoundedRect(x, y, w, h int, r float64) bool {
	fx, fy := float64(x)+0.5, float64(y)+0.5
	fw, fh := float64(w), float64(h)

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
	return dx*dx+dy*dy <= r*r
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
