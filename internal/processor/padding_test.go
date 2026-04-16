package processor

import (
	"image"
	"image/color"
	"testing"
)

func TestPad_ZeroRatio(t *testing.T) {
	src := newTestImage(64, 64)
	result := Pad(src, 0, nil)
	if result.Bounds().Dx() != 64 || result.Bounds().Dy() != 64 {
		t.Error("zero padding should return same size")
	}
}

func TestPad_NegativeRatio(t *testing.T) {
	src := newTestImage(64, 64)
	result := Pad(src, -0.1, nil)
	if result.Bounds().Dx() != 64 {
		t.Error("negative padding should return same size")
	}
}

func TestPad_WithRatio(t *testing.T) {
	src := newTestImage(80, 80)
	result := Pad(src, 0.1, nil)
	bounds := result.Bounds()
	// With 0.1 padding on 80px: padX = 80*0.1/(1-0.2) = 10, newW = 80+20 = 100
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestPad_At50PercentReturnsOriginal(t *testing.T) {
	src := newTestImage(64, 64)
	result := Pad(src, 0.5, nil)
	if result.Bounds().Dx() != 64 || result.Bounds().Dy() != 64 {
		t.Error("ratio >= 0.5 should return original image unchanged")
	}
}

func TestPad_Above50PercentReturnsOriginal(t *testing.T) {
	src := newTestImage(64, 64)
	result := Pad(src, 0.9, nil)
	if result.Bounds().Dx() != 64 || result.Bounds().Dy() != 64 {
		t.Error("ratio >= 0.5 should return original image unchanged")
	}
}

func TestPad_TransparentBackground(t *testing.T) {
	src := newTestImage(64, 64)
	result := Pad(src, 0.1, nil)
	// Corner should be transparent
	_, _, _, a := result.(image.Image).At(0, 0).RGBA()
	if a != 0 {
		t.Error("padding area with nil bg should be transparent")
	}
}

func TestPad_ColoredBackground(t *testing.T) {
	src := newTestImage(80, 80)
	bg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	result := Pad(src, 0.1, bg)
	// Corner pixel should be white
	r, g, b, a := result.(image.Image).At(0, 0).RGBA()
	if a == 0 {
		t.Error("padding area with white bg should be opaque")
	}
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("expected white, got r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

func TestPad_CenterPixelPreserved(t *testing.T) {
	src := newTestImage(80, 80) // red image
	result := Pad(src, 0.1, nil)
	bounds := result.(image.Image).Bounds()
	cx, cy := bounds.Dx()/2, bounds.Dy()/2
	r, _, _, a := result.(image.Image).At(cx, cy).RGBA()
	if a == 0 {
		t.Error("center pixel should be opaque")
	}
	if r>>8 != 255 {
		t.Errorf("center pixel should be red, got r=%d", r>>8)
	}
}
