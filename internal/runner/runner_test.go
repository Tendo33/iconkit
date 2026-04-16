package runner

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
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

func createTestJPEG(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: 120, G: 90, B: 60, A: 255})
		}
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
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

	f, _ := os.Open(results[0].Path)
	defer f.Close()
	img, _ := png.Decode(f)
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Error("corner pixel should be transparent after rounding")
	}
}

func TestRun_OriginalSizeOutput_WithRadiusOnly(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:              inputPath,
		Radius:             20,
		OriginalSizeOutput: true,
		Out:                outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if got := filepath.Base(results[0].Path); got != "icon.png" {
		t.Fatalf("output filename = %q, want %q", got, "icon.png")
	}

	if results[0].Size != 256 {
		t.Fatalf("result size = %d, want 256", results[0].Size)
	}

	f, err := os.Open(results[0].Path)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if img.Bounds().Dx() != 256 || img.Bounds().Dy() != 256 {
		t.Fatalf("output dimensions = %dx%d, want 256x256", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestRun_OriginalSizeOutput_WithPaddingOnly(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:              inputPath,
		Padding:            0.1,
		OriginalSizeOutput: true,
		Out:                outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if got := filepath.Base(results[0].Path); got != "icon.png" {
		t.Fatalf("output filename = %q, want %q", got, "icon.png")
	}
	if results[0].Size != 256 {
		t.Fatalf("result size = %d, want 256", results[0].Size)
	}
}

func TestRun_OriginalSizeOutput_WithBackgroundOnly(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "transparent.png")
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	f, _ := os.Create(imgPath)
	png.Encode(f, img)
	f.Close()

	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:              imgPath,
		BgColor:            color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		OriginalSizeOutput: true,
		Out:                outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	outF, err := os.Open(results[0].Path)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer outF.Close()

	outImg, err := png.Decode(outF)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if outImg.Bounds().Dx() != 256 || outImg.Bounds().Dy() != 256 {
		t.Fatalf("output dimensions = %dx%d, want 256x256", outImg.Bounds().Dx(), outImg.Bounds().Dy())
	}
	r, _, _, a := outImg.At(128, 128).RGBA()
	if a == 0 {
		t.Fatal("background should fill transparent pixels")
	}
	if r>>8 != 255 {
		t.Fatalf("background should be red, got r=%d", r>>8)
	}
}

func TestRun_OriginalSizeOutput_PreservesRectangularDimensions(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "banner.png", 320, 180)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:              inputPath,
		Radius:             20,
		OriginalSizeOutput: true,
		Out:                outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	f, err := os.Open(results[0].Path)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 180 {
		t.Fatalf("output dimensions = %dx%d, want 320x180", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestRun_WebPreset(t *testing.T) {
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

func TestRun_ChromeExtPreset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Preset: "chrome-ext",
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("chrome-ext preset should produce 4 icons, got %d", len(results))
	}
}

func TestRun_FirefoxExtPreset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Preset: "firefox-ext",
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("firefox-ext preset should produce 5 icons, got %d", len(results))
	}
}

func TestRun_PWAPreset(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 512, 512)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:  inputPath,
		Preset: "pwa",
		Out:    outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("pwa preset should produce 2 icons, got %d", len(results))
	}
}

func TestRun_WithPadding(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:   inputPath,
		Sizes:   []int{64},
		Padding: 0.1,
		Out:     outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	f, _ := os.Open(results[0].Path)
	defer f.Close()
	img, _ := png.Decode(f)
	if img.Bounds().Dx() != 64 {
		t.Errorf("output should still be 64px, got %d", img.Bounds().Dx())
	}
}

func TestRun_WithBackground(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a transparent PNG
	imgPath := filepath.Join(tmpDir, "transparent.png")
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	f, _ := os.Create(imgPath)
	png.Encode(f, img)
	f.Close()

	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:   imgPath,
		Sizes:   []int{32},
		BgColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		Out:     outDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}

	outF, _ := os.Open(results[0].Path)
	defer outF.Close()
	outImg, _ := png.Decode(outF)
	r, _, _, a := outImg.At(16, 16).RGBA()
	if a == 0 {
		t.Error("background should fill transparent pixels")
	}
	if r>>8 != 255 {
		t.Errorf("background should be red, got r=%d", r>>8)
	}
}

func TestRun_WithICO(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input: inputPath,
		Sizes: []int{16, 32, 48},
		Ico:   true,
		Out:   outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 PNGs + 1 ICO = 4
	if len(results) != 4 {
		t.Errorf("expected 4 results (3 png + 1 ico), got %d", len(results))
	}

	icoPath := filepath.Join(outDir, "favicon.ico")
	info, err := os.Stat(icoPath)
	if os.IsNotExist(err) {
		t.Fatal("favicon.ico should exist")
	}
	if info.Size() < 22 {
		t.Error("favicon.ico is too small")
	}
}

func TestRun_ICO_SkipsLargeSizes(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 512, 512)
	outDir := filepath.Join(tmpDir, "out")

	var buf bytes.Buffer
	results, err := Run(Options{
		Input: inputPath,
		Sizes: []int{16, 32, 512},
		Ico:   true,
		Out:   outDir,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	// 3 PNGs + 1 ICO (containing only 16+32) = 4
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
	output := buf.String()
	if !strings.Contains(output, "favicon") {
		t.Error("output should mention favicon.ico")
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

	_, err := Run(Options{Input: inputPath, Sizes: []int{32}, Out: outDir}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Run(Options{Input: inputPath, Sizes: []int{32}, Out: outDir}, nil)
	if err == nil {
		t.Error("expected error when file exists without force")
	}

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
	if len(results) != 2 {
		t.Errorf("expected 2 results for batch dir, got %d", len(results))
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

func TestRun_PaddingPlusBg(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:   inputPath,
		Sizes:   []int{64},
		Padding: 0.1,
		BgColor: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Out:     outDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}

	f, _ := os.Open(results[0].Path)
	defer f.Close()
	img, _ := png.Decode(f)
	// Corner should be white (bg color from padding)
	r, g, b, a := img.At(0, 0).RGBA()
	if a == 0 {
		t.Error("corner should be opaque with bg color")
	}
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("corner should be white, got r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

func TestRun_PaddingBgAndLargeRadiusKeepsOpaqueCorners(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := createTestPNG(t, tmpDir, "icon.png", 256, 256)
	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:   inputPath,
		Sizes:   []int{64},
		Padding: 0.1,
		Radius:  80,
		BgColor: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Out:     outDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}

	f, _ := os.Open(results[0].Path)
	defer f.Close()
	img, _ := png.Decode(f)

	r, g, b, a := img.At(0, 0).RGBA()
	if a == 0 {
		t.Fatal("corner should remain opaque when bg color is set")
	}
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Fatalf("corner should remain white, got r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

func TestRun_BatchWithICO(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	os.MkdirAll(inputDir, 0o755)
	createTestPNG(t, inputDir, "logo.png", 128, 128)

	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input: inputDir,
		Sizes: []int{16, 32},
		Ico:   true,
		Out:   outDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// 2 PNGs + 1 ICO = 3
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// ICO should be named logo.ico for batch mode
	icoPath := filepath.Join(outDir, "logo.ico")
	if _, err := os.Stat(icoPath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", icoPath)
	}
}

func TestRun_OriginalSizeOutput_BatchAvoidsNameCollisions(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	createTestPNG(t, inputDir, "logo.png", 128, 128)
	createTestJPEG(t, inputDir, "logo.jpg", 128, 128)

	outDir := filepath.Join(tmpDir, "out")

	results, err := Run(Options{
		Input:              inputDir,
		Radius:             18,
		OriginalSizeOutput: true,
		Out:                outDir,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	expected := map[string]bool{
		"logo-png.png": false,
		"logo-jpg.png": false,
	}
	for _, result := range results {
		name := filepath.Base(result.Path)
		if _, ok := expected[name]; !ok {
			t.Fatalf("unexpected output filename: %s", name)
		}
		expected[name] = true
	}

	for name, seen := range expected {
		if !seen {
			t.Fatalf("expected output file missing: %s", name)
		}
	}
}
