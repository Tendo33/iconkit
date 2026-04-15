package runner

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	"github.com/disintegration/imaging"
	"github.com/tudou/iconkit/internal/preset"
	"github.com/tudou/iconkit/internal/processor"
)

type Options struct {
	Input  string
	Sizes  []int
	Radius int
	Preset string
	Out    string
	Force  bool
}

var DefaultSizes = []int{16, 32, 64, 128}

type Result struct {
	Path string
	Size int
}

// Run executes the icon processing pipeline.
// writer receives progress messages; pass nil to suppress output.
func Run(opts Options, w io.Writer) ([]Result, error) {
	if w == nil {
		w = io.Discard
	}

	inputs, err := resolveInputs(opts.Input)
	if err != nil {
		return nil, err
	}

	sizes, err := resolveSizes(opts)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(opts.Out, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var results []Result
	for _, inputPath := range inputs {
		r, err := processOne(inputPath, sizes, opts, w)
		if err != nil {
			return results, err
		}
		results = append(results, r...)
	}

	return results, nil
}

func resolveInputs(input string) ([]string, error) {
	info, err := os.Stat(input)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("input file not found: %s", input)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access input: %w", err)
	}

	if !info.IsDir() {
		return []string{input}, nil
	}

	entries, err := os.ReadDir(input)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
			paths = append(paths, filepath.Join(input, e.Name()))
		}
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no image files found in directory: %s", input)
	}
	return paths, nil
}

func resolveSizes(opts Options) ([]int, error) {
	if opts.Preset != "" {
		p, ok := preset.Get(opts.Preset)
		if !ok {
			return nil, fmt.Errorf("unknown preset: %s (available: %s)",
				opts.Preset, strings.Join(preset.Names(), ", "))
		}
		return p.Sizes, nil
	}

	if len(opts.Sizes) > 0 {
		for _, s := range opts.Sizes {
			if s <= 0 {
				return nil, fmt.Errorf("invalid size: %d (must be > 0)", s)
			}
		}
		return opts.Sizes, nil
	}

	return DefaultSizes, nil
}

func processOne(inputPath string, sizes []int, opts Options, w io.Writer) ([]Result, error) {
	img, err := openImage(inputPath)
	if err != nil {
		return nil, err
	}

	originalSize := img.Bounds().Dx()
	baseName := fileBaseName(inputPath)
	multiFile := false
	if info, _ := os.Stat(opts.Input); info != nil && info.IsDir() {
		multiFile = true
	}

	var results []Result
	for _, size := range sizes {
		resized := processor.Resize(img, size)

		var output image.Image = resized
		if opts.Radius > 0 {
			scaledR := processor.ScaleRadius(opts.Radius, originalSize, size)
			output = processor.RoundCorners(resized, scaledR)
		}

		var filename string
		if multiFile {
			filename = fmt.Sprintf("%s-%d.png", baseName, size)
		} else {
			filename = fmt.Sprintf("icon-%d.png", size)
		}
		outPath := filepath.Join(opts.Out, filename)

		if !opts.Force {
			if _, err := os.Stat(outPath); err == nil {
				return results, fmt.Errorf("file already exists: %s (use -f to overwrite)", outPath)
			}
		}

		if err := savePNG(output, outPath); err != nil {
			return results, fmt.Errorf("failed to save %s: %w", outPath, err)
		}

		fmt.Fprintf(w, "  ✓ %s (%dx%d)\n", outPath, size, size)
		results = append(results, Result{Path: outPath, Size: size})
	}

	return results, nil
}

func openImage(path string) (image.Image, error) {
	img, err := imaging.Open(path)
	if err != nil {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".png", ".jpg", ".jpeg":
			return nil, fmt.Errorf("failed to open image: %w", err)
		default:
			return nil, fmt.Errorf("unsupported image format: %s", ext)
		}
	}
	return img, nil
}

func savePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fileBaseName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
