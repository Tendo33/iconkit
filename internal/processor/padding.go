package processor

import (
	"image"
	"image/color"
)

// Pad adds padding around the image as a ratio of the output size (0.0–0.5).
// The original image is centered; padding area is filled with bgColor.
func Pad(img image.Image, ratio float64, bgColor color.Color) image.Image {
	if ratio <= 0 {
		return img
	}
	if ratio >= 0.5 {
		return img
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	padX := int(float64(w) * ratio / (1 - 2*ratio))
	padY := int(float64(h) * ratio / (1 - 2*ratio))

	newW := w + 2*padX
	newH := h + 2*padY

	dst := image.NewNRGBA(image.Rect(0, 0, newW, newH))

	if bgColor == nil {
		bgColor = color.Transparent
	}
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			dst.Set(x, y, bgColor)
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x+padX, y+padY, img.At(x+bounds.Min.X, y+bounds.Min.Y))
		}
	}

	return dst
}
