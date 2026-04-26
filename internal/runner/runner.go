package runner

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "image/jpeg"

	"github.com/Tendo33/iconkit/internal/preset"
	"github.com/Tendo33/iconkit/internal/processor"
	"github.com/disintegration/imaging"
)

type Options struct {
	Input              string
	Sizes              []int
	Radius             int
	RadiusPercent      float64
	Preset             string
	Out                string
	Force              bool
	Padding            float64
	BgColor            color.Color // nil = transparent
	Ico                bool        // generate favicon.ico
	OriginalSizeOutput bool
	ResizeMode         string  // stretch (default), fit, cover
	OutputNameTemplate string  // template with {name},{size},{width},{height},{ext}
	GenerateHTML       bool
	GenerateManifest   bool
	Maskable           bool
	DryRun             bool
	Quiet              bool
	Verbose            bool
	ContinueOnError    bool
	Concurrency        int
	Format             string  // png (default), webp
	WebPQuality        float64 // 0-100, default 90
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
	if opts.Format == "" {
		opts.Format = "png"
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = runtime.NumCPU()
	}
	if opts.WebPQuality <= 0 {
		opts.WebPQuality = 90
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

	if !opts.DryRun {
		if err := os.MkdirAll(opts.Out, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	isMultiFile := len(inputs) > 1 || func() bool {
		info, _ := os.Stat(opts.Input)
		return info != nil && info.IsDir()
	}()

	// Show progress for large batches
	showProgress := len(inputs) > 5 && !opts.Quiet

	var mu sync.Mutex // guards writer and results
	results := make([]Result, 0)
	errors := make([]error, 0)

	type job struct {
		index     int
		inputPath string
	}
	type jobResult struct {
		index   int
		results []Result
		err     error
	}

	jobs := make(chan job, len(inputs))
	jobResults := make(chan jobResult, len(inputs))

	concurrency := opts.Concurrency
	if concurrency > len(inputs) {
		concurrency = len(inputs)
	}
	if concurrency < 1 {
		concurrency = 1
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				var localWriter io.Writer
				if opts.Quiet {
					localWriter = io.Discard
				} else {
					localWriter = &prefixWriter{mu: &mu, w: w}
				}
				r, e := processOne(j.inputPath, sizes, opts, outputNames[j.inputPath], localWriter, isMultiFile)
				jobResults <- jobResult{index: j.index, results: r, err: e}
			}
		}()
	}

	// Feed jobs
	for i, inputPath := range inputs {
		jobs <- job{index: i, inputPath: inputPath}
	}
	close(jobs)

	// Wait for all workers and close results channel
	go func() {
		wg.Wait()
		close(jobResults)
	}()

	// Ordered result collection
	resultMap := make(map[int][]Result, len(inputs))
	completed := 0
	for jr := range jobResults {
		completed++
		if showProgress {
			mu.Lock()
			fmt.Fprintf(w, "\rProcessing [%d/%d]...", completed, len(inputs))
			mu.Unlock()
		}
		if jr.err != nil {
			if opts.ContinueOnError {
				mu.Lock()
				errors = append(errors, jr.err)
				mu.Unlock()
			} else {
				// Drain remaining results
				for range jobResults {
				}
				return nil, jr.err
			}
		}
		resultMap[jr.index] = jr.results
	}

	if showProgress {
		fmt.Fprintln(w) // newline after progress
	}

	// Collect results in input order
	for i := range inputs {
		results = append(results, resultMap[i]...)
	}

	// Post-processing: HTML and manifest generation
	if opts.GenerateHTML && !opts.DryRun {
		if err := generateHTML(results, opts); err != nil {
			return results, fmt.Errorf("failed to generate HTML: %w", err)
		}
		if !opts.Quiet {
			fmt.Fprintf(w, "  ok %s (HTML link tags)\n", filepath.Join(opts.Out, "icons.html"))
		}
	}
	if opts.GenerateManifest && !opts.DryRun {
		if err := generateManifest(results, opts); err != nil {
			return results, fmt.Errorf("failed to generate manifest: %w", err)
		}
		if !opts.Quiet {
			fmt.Fprintf(w, "  ok %s (Web App Manifest)\n", filepath.Join(opts.Out, "manifest.json"))
		}
	}

	if len(errors) > 0 {
		fmt.Fprintln(w, "\nErrors encountered:")
		for _, e := range errors {
			fmt.Fprintf(w, "  - %s\n", e)
		}
		return results, fmt.Errorf("%d file(s) failed to process", len(errors))
	}

	return results, nil
}

// prefixWriter is a thread-safe writer that writes to a shared writer under a mutex.
type prefixWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.w.Write(b)
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
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".svg" {
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

	ext := outputExt(opts.Format)

	info, err := os.Stat(opts.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to access input: %w", err)
	}

	if !info.IsDir() {
		if len(inputs) == 1 {
			names[inputs[0]] = fmt.Sprintf("%s%s", fileBaseName(inputs[0]), ext)
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
			names[inputPath] = fmt.Sprintf("%s%s", baseName, ext)
			continue
		}
		names[inputPath] = fmt.Sprintf("%s-%s%s", baseName, normalizedInputExt(inputPath), ext)
	}

	return names, nil
}

func processOne(inputPath string, sizes []int, opts Options, originalOutputName string, w io.Writer, multiFile bool) ([]Result, error) {
	start := time.Now()

	img, err := openImage(inputPath, sizes)
	if err != nil {
		return nil, err
	}

	sourceWidth := img.Bounds().Dx()
	sourceHeight := img.Bounds().Dy()
	sourceScaleBase := sourceWidth
	if sourceHeight < sourceScaleBase {
		sourceScaleBase = sourceHeight
	}

	// Upscale warning
	if !opts.Quiet {
		maxTarget := 0
		for _, s := range sizes {
			if s > maxTarget {
				maxTarget = s
			}
		}
		if maxTarget > sourceScaleBase*2 && maxTarget > 0 {
			fmt.Fprintf(os.Stderr, "  warning: upscaling %s (%dpx) → %dpx, quality may degrade\n",
				filepath.Base(inputPath), sourceScaleBase, maxTarget)
		}
	}

	// Apply maskable padding first if --maskable
	paddingRatio := opts.Padding
	if opts.Maskable && paddingRatio < 0.18 {
		paddingRatio = 0.18
	}

	if paddingRatio > 0 {
		img = processor.Pad(img, paddingRatio, opts.BgColor)
	}

	baseName := fileBaseName(inputPath)

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

		var resized image.Image
		switch opts.ResizeMode {
		case "fit":
			resized = imaging.Fit(img, outputWidth, outputHeight, imaging.Lanczos)
			// Letterbox to exact target size
			bg := image.NewNRGBA(image.Rect(0, 0, outputWidth, outputHeight))
			if opts.BgColor != nil {
				r, g, b, _ := opts.BgColor.RGBA()
				for y := 0; y < outputHeight; y++ {
					for x := 0; x < outputWidth; x++ {
						bg.Set(x, y, color.NRGBA{
							R: uint8(r >> 8),
							G: uint8(g >> 8),
							B: uint8(b >> 8),
							A: 0xff,
						})
					}
				}
			}
			rBounds := resized.Bounds()
			offsetX := (outputWidth - rBounds.Dx()) / 2
			offsetY := (outputHeight - rBounds.Dy()) / 2
			for y := 0; y < rBounds.Dy(); y++ {
				for x := 0; x < rBounds.Dx(); x++ {
					bg.Set(x+offsetX, y+offsetY, resized.At(x, y))
				}
			}
			resized = bg
		case "cover":
			resized = imaging.Fill(img, outputWidth, outputHeight, imaging.Center, imaging.Lanczos)
		default: // stretch
			resized = imaging.Resize(img, outputWidth, outputHeight, imaging.Lanczos)
		}

		var output image.Image = resized

		// Determine effective radius
		effectiveRadius := opts.Radius
		if opts.RadiusPercent > 0 {
			minDim := outputWidth
			if outputHeight < minDim {
				minDim = outputHeight
			}
			effectiveRadius = int(float64(minDim) * opts.RadiusPercent / 100.0)
		} else if effectiveRadius > 0 && !opts.OriginalSizeOutput {
			effectiveRadius = processor.ScaleRadius(opts.Radius, sourceScaleBase, outputWidth)
		}

		if effectiveRadius > 0 {
			output = processor.RoundCorners(resized, effectiveRadius)
		}

		if opts.BgColor != nil {
			output = processor.FillBackground(output, opts.BgColor)
		}

		if opts.Ico && outputWidth == outputHeight && outputWidth <= 256 {
			icoImages = append(icoImages, output)
		}

		ext := outputExt(opts.Format)
		filename := outputFilename(baseName, outputWidth, outputHeight, ext, multiFile, opts.OriginalSizeOutput, originalOutputName, opts.OutputNameTemplate)
		outPath := filepath.Join(opts.Out, filename)

		if opts.DryRun {
			fmt.Fprintf(w, "  [dry-run] would write: %s (%dx%d)\n", outPath, outputWidth, outputHeight)
		} else {
			if !opts.Force {
				if _, err := os.Stat(outPath); err == nil {
					return results, fmt.Errorf("file already exists: %s (use -f to overwrite)", outPath)
				}
			}

			if err := saveImage(output, outPath, opts); err != nil {
				return results, fmt.Errorf("failed to save %s: %w", outPath, err)
			}

			if opts.Verbose {
				fmt.Fprintf(w, "  ok %s (%dx%d) [%s → %dx%d, %.0fms]\n",
					outPath, outputWidth, outputHeight,
					filepath.Base(inputPath), sourceWidth, sourceHeight,
					float64(time.Since(start).Milliseconds()))
			} else if !opts.Quiet {
				fmt.Fprintf(w, "  ok %s (%dx%d)\n", outPath, outputWidth, outputHeight)
			}
		}

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

		if opts.DryRun {
			fmt.Fprintf(w, "  [dry-run] would write: %s (favicon, %d sizes)\n", icoPath, len(icoImages))
		} else {
			if !opts.Force {
				if _, err := os.Stat(icoPath); err == nil {
					return results, fmt.Errorf("file already exists: %s (use -f to overwrite)", icoPath)
				}
			}

			if err := saveICO(icoImages, icoPath); err != nil {
				return results, fmt.Errorf("failed to save %s: %w", icoPath, err)
			}

			if !opts.Quiet {
				fmt.Fprintf(w, "  ok %s (favicon, %d sizes)\n", icoPath, len(icoImages))
			}
		}
		results = append(results, Result{Path: icoPath, Size: 0})
	}

	return results, nil
}

func openImage(path string, _ []int) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".svg":
		return processor.RasterizeSVG(path)
	case ".webp":
		return processor.DecodeWebP(path)
	case ".png", ".jpg", ".jpeg":
		img, err := imaging.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open image: %w", err)
		}
		return img, nil
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
}

func saveImage(img image.Image, path string, opts Options) error {
	if opts.Format == "webp" {
		return processor.EncodeWebP(img, path, float32(opts.WebPQuality))
	}
	return savePNG(img, path)
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

func outputExt(format string) string {
	if format == "webp" {
		return ".webp"
	}
	return ".png"
}

func fileBaseName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func normalizedInputExt(path string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
}

func outputFilename(baseName string, width, height int, ext string, multiFile bool, originalSizeOutput bool, originalOutputName string, tmpl string) string {
	if originalSizeOutput {
		if originalOutputName != "" {
			return originalOutputName
		}
		return baseName + ext
	}

	size := width // square; 0 for non-square
	if width != height {
		size = 0
	}

	if tmpl != "" {
		return applyNameTemplate(tmpl, baseName, size, width, height, ext)
	}

	if multiFile {
		return fmt.Sprintf("%s-%d%s", baseName, width, ext)
	}
	return fmt.Sprintf("icon-%d%s", width, ext)
}

func applyNameTemplate(tmpl, name string, size, width, height int, ext string) string {
	r := strings.NewReplacer(
		"{name}", name,
		"{size}", fmt.Sprintf("%d", size),
		"{width}", fmt.Sprintf("%d", width),
		"{height}", fmt.Sprintf("%d", height),
		"{ext}", strings.TrimPrefix(ext, "."),
	)
	result := r.Replace(tmpl)
	// Append ext if template doesn't include it
	if !strings.Contains(tmpl, "{ext}") && !strings.HasSuffix(result, ext) {
		result += ext
	}
	return result
}

// generateHTML writes an icons.html file with <link> tags for all generated PNG results.
func generateHTML(results []Result, opts Options) error {
	var sb strings.Builder
	for _, r := range results {
		if r.Size == 0 {
			continue
		}
		rel := "icon"
		if r.Size == 180 {
			rel = "apple-touch-icon"
		}
		// manifest.json and icons.html live in opts.Out, same dir as the icons
		filename := filepath.Base(r.Path)
		fmt.Fprintf(&sb, "<link rel=\"%s\" type=\"image/png\" sizes=\"%dx%d\" href=\"%s\">\n",
			rel, r.Width, r.Height, filename)
	}
	return os.WriteFile(filepath.Join(opts.Out, "icons.html"), []byte(sb.String()), 0o644)
}

// manifestIcon is a single icon entry in the Web App Manifest.
type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

// generateManifest writes a manifest.json for all generated results.
func generateManifest(results []Result, opts Options) error {
	mimeType := "image/png"
	if opts.Format == "webp" {
		mimeType = "image/webp"
	}

	icons := make([]manifestIcon, 0)
	for _, r := range results {
		if r.Size == 0 {
			continue
		}
		// manifest.json lives in opts.Out alongside the icons; use just the filename.
		filename := filepath.Base(r.Path)

		icon := manifestIcon{
			Src:   filename,
			Sizes: fmt.Sprintf("%dx%d", r.Width, r.Height),
			Type:  mimeType,
		}
		if opts.Maskable {
			icon.Purpose = "maskable"
		} else if r.Size == 512 {
			icon.Purpose = "any maskable"
		}
		icons = append(icons, icon)
	}

	manifest := struct {
		Icons []manifestIcon `json:"icons"`
	}{Icons: icons}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(opts.Out, "manifest.json"), data, 0o644)
}
