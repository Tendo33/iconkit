package processor

import (
	"image"
	"image/color"
	"testing"
)

func TestParseHexColor_6Digit(t *testing.T) {
	c, err := ParseHexColor("#ff0000")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("expected red, got %+v", c)
	}
}

func TestParseHexColor_NoHash(t *testing.T) {
	c, err := ParseHexColor("00ff00")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 0 || c.G != 255 || c.B != 0 {
		t.Errorf("expected green, got %+v", c)
	}
}

func TestParseHexColor_8Digit(t *testing.T) {
	c, err := ParseHexColor("#ff000080")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.A != 128 {
		t.Errorf("expected red with 50%% alpha, got %+v", c)
	}
}

func TestParseHexColor_3Digit(t *testing.T) {
	c, err := ParseHexColor("#f00")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.G != 0 || c.B != 0 {
		t.Errorf("expected red, got %+v", c)
	}
}

func TestParseHexColor_Invalid(t *testing.T) {
	invalid := []string{"xyz", "#gg0000", "#1", "#12345", ""}
	for _, s := range invalid {
		_, err := ParseHexColor(s)
		if err == nil {
			t.Errorf("expected error for %q", s)
		}
	}
}

func TestFillBackground_OpaqueImage(t *testing.T) {
	src := newTestImage(32, 32) // fully opaque red
	bg := color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	result := FillBackground(src, bg)

	// Should remain red (source is fully opaque)
	r, _, b, _ := result.At(16, 16).RGBA()
	if r>>8 != 255 || b>>8 != 0 {
		t.Error("opaque source pixel should remain unchanged")
	}
}

func TestFillBackground_TransparentImage(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	// Leave all pixels transparent
	bg := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	result := FillBackground(src, bg)

	// Should be green
	_, g, _, a := result.At(16, 16).RGBA()
	if g>>8 != 255 || a == 0 {
		t.Error("transparent pixel should become background color")
	}
}

func TestFillBackground_SemiTransparent(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 128})
	bg := color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	result := FillBackground(src, bg)

	r, _, b, a := result.At(0, 0).RGBA()
	if a == 0 {
		t.Error("composited pixel should be opaque")
	}
	rr := r >> 8
	bb := b >> 8
	if rr < 100 || rr > 160 {
		t.Errorf("red channel should be ~128 (50%% blend), got %d", rr)
	}
	if bb < 95 || bb > 160 {
		t.Errorf("blue channel should be ~127 (50%% blend), got %d", bb)
	}
}
