package processor

import (
	"bytes"
	"image"
	"testing"
)

func TestEncodeICO_Single(t *testing.T) {
	img := newTestImage(32, 32)
	var buf bytes.Buffer
	err := EncodeICO(&buf, []image.Image{img})
	if err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	// ICO header: 6 bytes, 1 entry: 16 bytes, total header = 22 bytes minimum
	if len(data) < 22 {
		t.Errorf("ICO too small: %d bytes", len(data))
	}
	// Check magic: reserved=0, type=1 (ICO), count=1
	if data[0] != 0 || data[1] != 0 { // reserved
		t.Error("reserved should be 0")
	}
	if data[2] != 1 || data[3] != 0 { // type = 1
		t.Error("type should be 1 (ICO)")
	}
	if data[4] != 1 || data[5] != 0 { // count = 1
		t.Error("count should be 1")
	}
}

func TestEncodeICO_Multiple(t *testing.T) {
	images := []image.Image{
		newTestImage(16, 16),
		newTestImage(32, 32),
		newTestImage(48, 48),
	}
	var buf bytes.Buffer
	err := EncodeICO(&buf, images)
	if err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	// count = 3
	if data[4] != 3 || data[5] != 0 {
		t.Errorf("count should be 3, got %d", int(data[4])+int(data[5])<<8)
	}
}

func TestEncodeICO_256(t *testing.T) {
	img := newTestImage(256, 256)
	var buf bytes.Buffer
	err := EncodeICO(&buf, []image.Image{img})
	if err != nil {
		t.Fatal(err)
	}
	// Width/height should be 0 for 256
	// Entry starts at byte 6
	if buf.Bytes()[6] != 0 || buf.Bytes()[7] != 0 {
		t.Error("256px should be encoded as 0 in ICO entry")
	}
}

func TestEncodeICO_SkipsOversized(t *testing.T) {
	images := []image.Image{
		newTestImage(32, 32),
		newTestImage(512, 512), // should be skipped
	}
	var buf bytes.Buffer
	err := EncodeICO(&buf, images)
	if err != nil {
		t.Fatal(err)
	}
	// count should be 1 (512 skipped)
	if buf.Bytes()[4] != 1 {
		t.Errorf("expected 1 entry (512 skipped), got %d", buf.Bytes()[4])
	}
}

func TestEncodeICO_AllOversized(t *testing.T) {
	images := []image.Image{
		newTestImage(512, 512),
	}
	var buf bytes.Buffer
	err := EncodeICO(&buf, images)
	if err == nil {
		t.Error("expected error when all images are too large")
	}
}
