package processor

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

// FillBackground composites the image onto a solid background color.
// Transparent pixels become the background color.
func FillBackground(img image.Image, bgColor color.Color) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, w, h))

	bgR, bgG, bgB, bgA := bgColor.RGBA()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src := img.At(x+bounds.Min.X, y+bounds.Min.Y)
			sR, sG, sB, sA := src.RGBA()

		if sA == 0xffff {
			dst.Set(x, y, src)
		} else if sA == 0 {
			dst.Set(x, y, bgColor)
		} else {
			invA := 0xffff - sA
			outR := sR + bgR*invA/0xffff
			outG := sG + bgG*invA/0xffff
			outB := sB + bgB*invA/0xffff
			outA := sA + bgA*invA/0xffff

			var nr, ng, nb uint8
			if outA > 0 {
				nr = uint8((outR * 0xff / outA) & 0xff)
				ng = uint8((outG * 0xff / outA) & 0xff)
				nb = uint8((outB * 0xff / outA) & 0xff)
			}
			dst.Set(x, y, color.NRGBA{
				R: nr,
				G: ng,
				B: nb,
				A: uint8(outA >> 8),
			})
		}
		}
	}

	return dst
}

// ParseHexColor parses a hex color string like "#ff0000" or "ff0000".
func ParseHexColor(s string) (color.NRGBA, error) {
	s = strings.TrimPrefix(s, "#")

	var r, g, b, a uint8
	a = 255

	switch len(s) {
	case 6: // RRGGBB
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color: #%s", s)
		}
	case 8: // RRGGBBAA
		_, err := fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color: #%s", s)
		}
	case 3: // RGB shorthand
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color: #%s", s)
		}
		r = r*16 + r
		g = g*16 + g
		b = b*16 + b
	default:
		return color.NRGBA{}, fmt.Errorf("invalid hex color: #%s (use 3, 6, or 8 hex digits)", s)
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}, nil
}
