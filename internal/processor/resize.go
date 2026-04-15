package processor

import (
	"image"

	"github.com/disintegration/imaging"
)

func Resize(img image.Image, size int) image.Image {
	return imaging.Resize(img, size, size, imaging.Lanczos)
}
