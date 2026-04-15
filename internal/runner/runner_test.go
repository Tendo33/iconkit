package runner

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func createTestPNG(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRun_DefaultSizes(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	var buf bytes.Buffer
	results, err := Run(Options{
		Input: inputPath,
		Out:   outDir,
	}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != len(DefaultSizes) {
		t.Errorf("expected %d results, got %d", len(DefaultSizes), len(results))
	}

	for _, r := range results {
		if _, err := os.Stat(r.Path); os.IsNotExist(err) {
			t.Errorf("output file does not exist: %s", r.Path)
		}
	}
}

func TestRun_CustomSizes(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input: inputPath,
		Sizes: []int{48, 96},
		Out:   outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	for _, expected := range []string{"icon-48.png", "icon-96.png"} {
		path := filepath.Join(outDir, expected)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", expected)
		}
	}
}

func TestRun_WithRadius(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Sizes:  []int{64},
		Radius: 20,
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Verify the corner pixel is transparent
	f, _ := os.Open(results[0].Path)
	defer f.Close()
	img, _ := png.Decode(f)
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Error("corner pixel should be transparent after rounding")
	}
}

func TestRun_Preset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 512, 512)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Preset: "web",
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 6 {
		t.Errorf("web preset should produce 6 icons, got %d", len(results))
	}
}

func TestRun_iOSPreset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 1024, 1024)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Preset: "ios",
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 13 {
		t.Errorf("ios preset should produce 13 icons, got %d", len(results))
	}
}

func TestRun_AndroidPreset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 512, 512)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Preset: "android",
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 7 {
		t.Errorf("android preset should produce 7 icons, got %d", len(results))
	}
}

func TestRun_FileNotFound(t *testing.T) {
	_, err := Run(Options{
		Input: "/nonexistent/icon.png",
		Out:   t.TempDir(),
	}, nil)
	if err == nil {
		t.Error("expected error for nonexistent input")
	}
}

func TestRun_UnknownPreset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 64, 64)

	_, err := Run(Options{
		Input:  inputPath,
		Preset: "unknown",
		Out:    filepath.Join(tmpDir, "out"),
	}, nil)
	if err == nil {
		t.Error("expected error for unknown preset")
	}
}

func TestRun_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 64, 64)
	outDir := filepath.Join(tmpDir, "out")

	// First run
	_, err := Run(Options{Input: inputPath, Sizes: []int{32}, Out: outDir}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Second run without force should fail
	_, err = Run(Options{Input: inputPath, Sizes: []int{32}, Out: outDir}, nil)
	if err == nil {
		t.Error("expected error when file exists without force")
	}

	// Third run with force should succeed
	_, err = Run(Options{Input: inputPath, Sizes: []int{32}, Out: outDir, Force: true}, nil)
	if err != nil {
		t.Errorf("force overwrite should succeed: %v", err)
	}
}

func TestRun_BatchDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	os.MkdirAll(inputDir, 0o755)
	createTestPNG(t, inputDir, "a.png", 64, 64)
	createTestPNG(t, inputDir, "b.png", 64, 64)

	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input: inputDir,
		Sizes: []int{32},
		Out:   outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 images × 1 size = 2 results
	if len(results) != 2 {
		t.Errorf("expected 2 results for batch dir, got %d", len(results))
	}

	// Check naming: {basename}-{size}.png
	for _, name := range []string{"a-32.png", "b-32.png"} {
		path := filepath.Join(outDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", name)
		}
	}
}

func TestRun_BatchDirectory_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(inputDir, 0o755)

	_, err := Run(Options{
		Input: inputDir,
		Out:   filepath.Join(tmpDir, "out"),
	}, nil)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestRun_InvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 64, 64)

	_, err := Run(Options{
		Input: inputPath,
		Sizes: []int{-1},
		Out:   filepath.Join(tmpDir, "out"),
	}, nil)
	if err == nil {
		t.Error("expected error for negative size")
	}
}

func TestRun_NilWriter(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 64, 64)
	outDir := filepath.Join(tmpDir, "out")

	// Should not panic with nil writer
	results, err := Run(Options{
		Input: inputPath,
		Sizes: []int{16},
		Out:   outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRun_OutputSizeCorrect(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 512, 512)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input: inputPath,
		Sizes: []int{64, 128},
		Out:   outDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range results {
		f, _ := os.Open(r.Path)
		img, _ := png.Decode(f)
		f.Close()
		bounds := img.Bounds()
		if bounds.Dx() != r.Size || bounds.Dy() != r.Size {
			t.Errorf("file %s: expected %dx%d, got %dx%d",
				r.Path, r.Size, r.Size, bounds.Dx(), bounds.Dy())
		}
	}
}
