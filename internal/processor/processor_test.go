package processor

import (
	"image"
	"image/color"
	"testing"
)

func newTestImage(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	return img
}

func TestResize(t *testing.T) {
	tests := []struct {
		name       string
		srcW, srcH int
		targetSize int
	}{
		{"downscale", 256, 256, 64},
		{"upscale", 16, 16, 128},
		{"same size", 64, 64, 64},
		{"small", 100, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := newTestImage(tt.srcW, tt.srcH)
			result := Resize(src, tt.targetSize)
			bounds := result.Bounds()
			if bounds.Dx() != tt.targetSize || bounds.Dy() != tt.targetSize {
				t.Errorf("expected %dx%d, got %dx%d",
					tt.targetSize, tt.targetSize, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestRoundCorners_ZeroRadius(t *testing.T) {
	src := newTestImage(64, 64)
	result := RoundCorners(src, 0)
	if result.Bounds().Dx() != 64 || result.Bounds().Dy() != 64 {
		t.Error("size should be preserved")
	}
	// Center pixel should remain opaque
	_, _, _, a := result.At(32, 32).RGBA()
	if a == 0 {
		t.Error("center pixel should be opaque with radius 0")
	}
}

func TestRoundCorners_WithRadius(t *testing.T) {
	src := newTestImage(100, 100)
	result := RoundCorners(src, 20)

	if result.Bounds().Dx() != 100 || result.Bounds().Dy() != 100 {
		t.Error("size should be preserved")
	}

	// Corner pixel (0,0) should be transparent
	_, _, _, a := result.At(0, 0).RGBA()
	if a != 0 {
		t.Errorf("corner pixel should be transparent, got alpha=%d", a)
	}

	// Center pixel should be opaque
	_, _, _, a = result.At(50, 50).RGBA()
	if a == 0 {
		t.Error("center pixel should be opaque")
	}
}

func TestRoundCorners_AllCorners(t *testing.T) {
	src := newTestImage(100, 100)
	result := RoundCorners(src, 30)

	corners := []struct {
		name string
		x, y int
	}{
		{"top-left", 0, 0},
		{"top-right", 99, 0},
		{"bottom-left", 0, 99},
		{"bottom-right", 99, 99},
	}

	for _, c := range corners {
		_, _, _, a := result.At(c.x, c.y).RGBA()
		if a != 0 {
			t.Errorf("%s corner should be transparent", c.name)
		}
	}
}

func TestRoundCorners_RadiusClamped(t *testing.T) {
	src := newTestImage(20, 20)
	// Radius larger than half the image should be clamped
	result := RoundCorners(src, 100)
	if result.Bounds().Dx() != 20 || result.Bounds().Dy() != 20 {
		t.Error("size should be preserved even with oversized radius")
	}
}

func TestScaleRadius(t *testing.T) {
	tests := []struct {
		name                                  string
		originalRadius, originalSize, target   int
		expected                               int
	}{
		{"proportional downscale", 20, 100, 50, 10},
		{"proportional upscale", 10, 50, 100, 20},
		{"same size", 15, 100, 100, 15},
		{"zero original size", 10, 0, 100, 0},
		{"zero radius", 0, 100, 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScaleRadius(tt.originalRadius, tt.originalSize, tt.target)
			if got != tt.expected {
				t.Errorf("ScaleRadius(%d, %d, %d) = %d, want %d",
					tt.originalRadius, tt.originalSize, tt.target, got, tt.expected)
			}
		})
	}
}

func TestInRoundedRect(t *testing.T) {
	// 100x100 rect with radius 20
	// Center should be inside
	if !inRoundedRect(50, 50, 100, 100, 20) {
		t.Error("center should be inside")
	}
	// Corner origin should be outside
	if inRoundedRect(0, 0, 100, 100, 20) {
		t.Error("corner (0,0) should be outside with r=20")
	}
	// Just inside the radius arc
	if !inRoundedRect(20, 20, 100, 100, 20) {
		t.Error("(20,20) should be inside")
	}
}
