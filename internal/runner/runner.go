package runner

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	"github.com/Tendo33/iconkit/internal/preset"
	"github.com/Tendo33/iconkit/internal/processor"
	"github.com/disintegration/imaging"
)

type Options struct {
	Input              string
	Sizes              []int
	Radius             int
	Preset             string
	Out                string
	Force              bool
	Padding            float64
	BgColor            color.Color // nil = transparent
	Ico                bool        // generate favicon.ico
	OriginalSizeOutput bool
}

var DefaultSizes = []int{16, 32, 64, 128}

type Result struct {
	Path   string
	Size   int
	Width  int
	Height int
}

func Run(opts Options, w io.Writer) ([]Result, error) {
	if w == nil {
		w = io.Discard
	}

	inputs, err := resolveInputs(opts.Input)
	if err != nil {
		return nil, err
	}

	outputNames, err := buildOriginalOutputNames(inputs, opts)
	if err != nil {
		return nil, err
	}

	var sizes []int
	if !opts.OriginalSizeOutput {
		sizes, err = resolveSizes(opts)
		if err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(opts.Out, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var results []Result
	for _, inputPath := range inputs {
		r, err := processOne(inputPath, sizes, opts, outputNames[inputPath], w)
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

func buildOriginalOutputNames(inputs []string, opts Options) (map[string]string, error) {
	names := make(map[string]string, len(inputs))
	if !opts.OriginalSizeOutput {
		return names, nil
	}

	info, err := os.Stat(opts.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to access input: %w", err)
	}

	if !info.IsDir() {
		if len(inputs) == 1 {
			names[inputs[0]] = fmt.Sprintf("%s.png", fileBaseName(inputs[0]))
		}
		return names, nil
	}

	baseNameCounts := make(map[string]int, len(inputs))
	for _, inputPath := range inputs {
		baseNameCounts[fileBaseName(inputPath)]++
	}

	for _, inputPath := range inputs {
		baseName := fileBaseName(inputPath)
		if baseNameCounts[baseName] == 1 {
			names[inputPath] = fmt.Sprintf("%s.png", baseName)
			continue
		}
		names[inputPath] = fmt.Sprintf("%s-%s.png", baseName, normalizedInputExt(inputPath))
	}

	return names, nil
}

func processOne(inputPath string, sizes []int, opts Options, originalOutputName string, w io.Writer) ([]Result, error) {
	img, err := openImage(inputPath)
	if err != nil {
		return nil, err
	}

	sourceWidth := img.Bounds().Dx()
	sourceHeight := img.Bounds().Dy()
	sourceScaleBase := sourceWidth
	if sourceHeight < sourceScaleBase {
		sourceScaleBase = sourceHeight
	}

	if opts.Padding > 0 {
		img = processor.Pad(img, opts.Padding, opts.BgColor)
	}

	baseName := fileBaseName(inputPath)
	multiFile := false
	if info, _ := os.Stat(opts.Input); info != nil && info.IsDir() {
		multiFile = true
	}

	type targetDimensions struct {
		width  int
		height int
	}

	targets := make([]targetDimensions, 0, len(sizes))
	if opts.OriginalSizeOutput {
		targets = append(targets, targetDimensions{width: sourceWidth, height: sourceHeight})
	} else {
		for _, size := range sizes {
			targets = append(targets, targetDimensions{width: size, height: size})
		}
	}

	var results []Result
	var icoImages []image.Image

	for _, target := range targets {
		outputWidth := target.width
		outputHeight := target.height
		resized := imaging.Resize(img, outputWidth, outputHeight, imaging.Lanczos)

		var output image.Image = resized
		if opts.Radius > 0 {
			scaledRadius := opts.Radius
			if !opts.OriginalSizeOutput {
				scaledRadius = processor.ScaleRadius(opts.Radius, sourceScaleBase, outputWidth)
			}
			output = processor.RoundCorners(resized, scaledRadius)
		}

		if opts.BgColor != nil {
			output = processor.FillBackground(output, opts.BgColor)
		}

		if opts.Ico && outputWidth == outputHeight && outputWidth <= 256 {
			icoImages = append(icoImages, output)
		}

		filename := outputFilename(baseName, outputWidth, multiFile, opts.OriginalSizeOutput, originalOutputName)
		outPath := filepath.Join(opts.Out, filename)

		if !opts.Force {
			if _, err := os.Stat(outPath); err == nil {
				return results, fmt.Errorf("file already exists: %s (use -f to overwrite)", outPath)
			}
		}

		if err := savePNG(output, outPath); err != nil {
			return results, fmt.Errorf("failed to save %s: %w", outPath, err)
		}

		fmt.Fprintf(w, "  ok %s (%dx%d)\n", outPath, outputWidth, outputHeight)

		resultSize := outputWidth
		if outputWidth != outputHeight {
			resultSize = 0
		}
		results = append(results, Result{
			Path:   outPath,
			Size:   resultSize,
			Width:  outputWidth,
			Height: outputHeight,
		})
	}

	if opts.Ico && len(icoImages) > 0 {
		icoName := "favicon.ico"
		if multiFile {
			icoName = fmt.Sprintf("%s.ico", baseName)
		}
		icoPath := filepath.Join(opts.Out, icoName)

		if !opts.Force {
			if _, err := os.Stat(icoPath); err == nil {
				return results, fmt.Errorf("file already exists: %s (use -f to overwrite)", icoPath)
			}
		}

		if err := saveICO(icoImages, icoPath); err != nil {
			return results, fmt.Errorf("failed to save %s: %w", icoPath, err)
		}

		fmt.Fprintf(w, "  ok %s (favicon, %d sizes)\n", icoPath, len(icoImages))
		results = append(results, Result{Path: icoPath, Size: 0})
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

func saveICO(images []image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return processor.EncodeICO(f, images)
}

func fileBaseName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func normalizedInputExt(path string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
}

func outputFilename(baseName string, size int, multiFile bool, originalSizeOutput bool, originalOutputName string) string {
	if originalSizeOutput {
		if originalOutputName != "" {
			return originalOutputName
		}
		return fmt.Sprintf("%s.png", baseName)
	}
	if multiFile {
		return fmt.Sprintf("%s-%d.png", baseName, size)
	}
	return fmt.Sprintf("icon-%d.png", size)
}
