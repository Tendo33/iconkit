package processor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
)

// ICO file format: https://en.wikipedia.org/wiki/ICO_(file_format)

type icoHeader struct {
	Reserved uint16
	Type     uint16
	Count    uint16
}

type icoDirEntry struct {
	Width       uint8
	Height      uint8
	ColorCount  uint8
	Reserved    uint8
	Planes      uint16
	BitCount    uint16
	BytesInRes  uint32
	ImageOffset uint32
}

// EncodeICO writes multiple images as a single .ico file.
// Each image is stored as an embedded PNG.
// Images larger than 256x256 are skipped (ICO spec limit).
func EncodeICO(w io.Writer, images []image.Image) error {
	var validImages []image.Image
	for _, img := range images {
		b := img.Bounds()
		if b.Dx() <= 256 && b.Dy() <= 256 {
			validImages = append(validImages, img)
		}
	}

	if len(validImages) == 0 {
		return fmt.Errorf("no images suitable for ICO (max 256x256)")
	}

	// Encode each image to PNG in memory
	pngData := make([][]byte, len(validImages))
	for i, img := range validImages {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return err
		}
		pngData[i] = buf.Bytes()
	}

	header := icoHeader{
		Reserved: 0,
		Type:     1, // ICO
		Count:    uint16(len(validImages)),
	}
	if err := binary.Write(w, binary.LittleEndian, header); err != nil {
		return err
	}

	// Calculate offsets: header (6 bytes) + entries (16 bytes each)
	dataOffset := uint32(6 + 16*len(validImages))

	for i, img := range validImages {
		b := img.Bounds()
		width := uint8(b.Dx())
		height := uint8(b.Dy())
		if b.Dx() == 256 {
			width = 0 // ICO spec: 0 means 256
		}
		if b.Dy() == 256 {
			height = 0
		}

		entry := icoDirEntry{
			Width:       width,
			Height:      height,
			ColorCount:  0,
			Reserved:    0,
			Planes:      1,
			BitCount:    32,
			BytesInRes:  uint32(len(pngData[i])),
			ImageOffset: dataOffset,
		}
		if err := binary.Write(w, binary.LittleEndian, entry); err != nil {
			return err
		}
		dataOffset += uint32(len(pngData[i]))
	}

	for _, data := range pngData {
		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	return nil
}
