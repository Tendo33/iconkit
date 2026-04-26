package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tendo33/iconkit/internal/config"
	"github.com/Tendo33/iconkit/internal/preset"
	"github.com/Tendo33/iconkit/internal/processor"
	"github.com/Tendo33/iconkit/internal/runner"
	"github.com/spf13/cobra"
)

var version = "dev"

var (
	sizes          string
	radius         int
	radiusPercent  float64
	presetName     string
	outDir         string
	force          bool
	configFile     string
	padding        float64
	bgColor        string
	ico            bool
	resizeMode     string
	outputName     string
	generateHTML   bool
	generateManifest bool
	maskable       bool
	dryRun         bool
	quiet          bool
	verbose        bool
	continueOnErr  bool
	concurrency    int
	format         string
	webpQuality    float64
)

var rootCmd = &cobra.Command{
	Use:   "iconkit [input]",
	Short: "Icon processing CLI - resize and round corners in one command",
	Long: `iconkit is a developer-friendly CLI tool for icon processing.

It takes a single image (or a directory of images) and outputs multiple sizes
with optional rounded corners, padding, background color, and favicon.ico.
Without -s or -p, using -r, --pad, or --bg keeps the original dimensions and
writes a single processed PNG to the output directory.
In batch mode, files are written as {name}.png, and same-name conflicts become {name}-{source-ext}.png.
When --ico is enabled, iconkit keeps the existing multi-size favicon flow.

Examples:
  iconkit icon.png
  iconkit icon.png -s 16,32,64,128
  iconkit icon.png -r 20
  iconkit icon.png --radius-percent 25
  iconkit icon.png -r 20 -s 16,32,64,128
  iconkit icon.png -p web
  iconkit icon.png -p chrome-ext
  iconkit icon.png -p firefox-ext
  iconkit icon.png --ico -p web
  iconkit icon.png --pad 0.1 --bg "#ffffff"
  iconkit icon.png --resize-mode fit --bg "#ffffff" -s 128
  iconkit icon.png -p web --html --manifest
  iconkit icon.png -p android --maskable
  iconkit -c iconkit.json
  iconkit icon.png -r 24 -s 16,32,64 -o ./dist
  iconkit ./assets/ -p web -j 4
  iconkit icon.png --dry-run -p ios`,
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	RunE:    run,
}

func init() {
	// Core flags
	rootCmd.Flags().StringVarP(&sizes, "sizes", "s", "", "output sizes, comma-separated; overrides processing-only mode (e.g. 16,32,64)")
	rootCmd.Flags().IntVarP(&radius, "radius", "r", 0, "corner radius in pixels; without -s/-p outputs one PNG at the original size")
	rootCmd.Flags().Float64Var(&radiusPercent, "radius-percent", 0, "corner radius as percent of min dimension (0-50); mutually exclusive with --radius")
	rootCmd.Flags().StringVarP(&presetName, "preset", "p", "", "size preset (web, ios, android, chrome-ext, firefox-ext, pwa, macos, windows, electron, tauri)")
	rootCmd.Flags().StringVarP(&outDir, "out", "o", "", "output directory (default: ./icons)")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path (default: auto-detect iconkit.json)")
	rootCmd.Flags().Float64Var(&padding, "pad", 0, "padding ratio around icon (0.0-0.5, e.g. 0.1 = 10%); without -s/-p outputs one PNG at the original size")
	rootCmd.Flags().StringVar(&bgColor, "bg", "", "background color in hex (e.g. \"#ffffff\", \"ff0000\"); without -s/-p outputs one PNG at the original size")
	rootCmd.Flags().BoolVar(&ico, "ico", false, "also generate favicon.ico (sizes <= 256); keeps multi-size output")

	// Resize mode
	rootCmd.Flags().StringVar(&resizeMode, "resize-mode", "stretch", "resize mode: stretch (default), fit (letterbox), cover (crop center)")

	// Output naming
	rootCmd.Flags().StringVar(&outputName, "output-name", "", "output filename template: {name},{size},{width},{height},{ext} (e.g. \"{width}x{height}\")")

	// Ecosystem generation
	rootCmd.Flags().BoolVar(&generateHTML, "html", false, "generate icons.html with <link> tags for all output icons")
	rootCmd.Flags().BoolVar(&generateManifest, "manifest", false, "generate manifest.json (Web App Manifest) for all output icons")
	rootCmd.Flags().BoolVar(&maskable, "maskable", false, "apply 18% padding for Android maskable icons (sets --pad 0.18 if not larger)")

	// DX flags
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview output without writing any files")
	rootCmd.Flags().BoolVar(&quiet, "quiet", false, "suppress per-file output, only print final summary")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "print per-file processing details including timing")
	rootCmd.Flags().BoolVar(&continueOnErr, "continue-on-error", false, "in batch mode, continue processing on per-file errors")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "j", 0, "number of parallel workers for batch processing (default: NumCPU)")

	// Format flags
	rootCmd.Flags().StringVar(&format, "format", "png", "output format: png (default), webp")
	rootCmd.Flags().Float64Var(&webpQuality, "webp-quality", 90, "WebP output quality (0-100, default 90)")
}

func Execute() error {
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	inputPath := ""
	if len(args) > 0 {
		inputPath = args[0]
	}

	opts, err := buildOptions(inputPath)
	if err != nil {
		return err
	}

	results, err := runner.Run(opts, os.Stdout)
	if err != nil && !opts.ContinueOnError {
		return err
	}

	if opts.DryRun {
		fmt.Printf("\n[dry-run] total: %d file(s) would be written to %s\n", len(results), opts.Out)
	} else if !opts.Quiet {
		fmt.Printf("\nDone! %d files saved to %s\n", len(results), opts.Out)
	} else {
		fmt.Printf("Done! %d files saved to %s\n", len(results), opts.Out)
	}
	return err
}

func buildOptions(inputPath string) (runner.Options, error) {
	opts := runner.Options{
		Input:            inputPath,
		Radius:           radius,
		RadiusPercent:    radiusPercent,
		Out:              "./icons",
		Force:            force,
		Padding:          padding,
		Ico:              ico,
		ResizeMode:       resizeMode,
		OutputNameTemplate: outputName,
		GenerateHTML:     generateHTML,
		GenerateManifest: generateManifest,
		Maskable:         maskable,
		DryRun:           dryRun,
		Quiet:            quiet,
		Verbose:          verbose,
		ContinueOnError:  continueOnErr,
		Concurrency:      concurrency,
		Format:           format,
		WebPQuality:      webpQuality,
	}

	// Validate mutually exclusive flags
	if radius > 0 && radiusPercent > 0 {
		return opts, fmt.Errorf("--radius and --radius-percent are mutually exclusive")
	}
	if quiet && verbose {
		return opts, fmt.Errorf("--quiet and --verbose are mutually exclusive")
	}
	if radiusPercent < 0 || radiusPercent > 50 {
		return opts, fmt.Errorf("--radius-percent must be between 0 and 50, got %v", radiusPercent)
	}

	// Load config file (explicit or auto-detect)
	var cfg *config.Config
	var err error
	if configFile != "" {
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			return opts, err
		}
	} else {
		cfg, _ = config.Load(".")
	}

	// Apply config as defaults (CLI flags override)
	if cfg != nil {
		if cfg.Input != "" && opts.Input == "" {
			opts.Input = cfg.Input
		}
		if cfg.Radius > 0 && radius == 0 {
			opts.Radius = cfg.Radius
		}
		if cfg.RadiusPercent > 0 && radiusPercent == 0 {
			opts.RadiusPercent = cfg.RadiusPercent
		}
		if cfg.Out != "" && outDir == "" {
			opts.Out = cfg.Out
		}
		if cfg.Force && !force {
			opts.Force = cfg.Force
		}
		if cfg.Padding > 0 && padding == 0 {
			opts.Padding = cfg.Padding
		}
		if cfg.Bg != "" && bgColor == "" {
			bgColor = cfg.Bg
		}
		if cfg.Ico && !ico {
			opts.Ico = cfg.Ico
		}
		if cfg.Preset != "" && presetName == "" {
			presetName = cfg.Preset
		}
		if len(cfg.Sizes) > 0 && sizes == "" {
			opts.Sizes = cfg.Sizes
		}
		if cfg.ResizeMode != "" && resizeMode == "stretch" {
			opts.ResizeMode = cfg.ResizeMode
		}
		if cfg.OutputNameTemplate != "" && outputName == "" {
			opts.OutputNameTemplate = cfg.OutputNameTemplate
		}
		if cfg.GenerateHTML && !generateHTML {
			opts.GenerateHTML = cfg.GenerateHTML
		}
		if cfg.GenerateManifest && !generateManifest {
			opts.GenerateManifest = cfg.GenerateManifest
		}
		if cfg.Maskable && !maskable {
			opts.Maskable = cfg.Maskable
		}
		if cfg.DryRun && !dryRun {
			opts.DryRun = cfg.DryRun
		}
		if cfg.Quiet && !quiet {
			opts.Quiet = cfg.Quiet
		}
		if cfg.Verbose && !verbose {
			opts.Verbose = cfg.Verbose
		}
		if cfg.ContinueOnError && !continueOnErr {
			opts.ContinueOnError = cfg.ContinueOnError
		}
		if cfg.Concurrency > 0 && concurrency == 0 {
			opts.Concurrency = cfg.Concurrency
		}
		if cfg.Format != "" && format == "png" {
			opts.Format = cfg.Format
		}
		if cfg.WebPQuality > 0 && webpQuality == 90 {
			opts.WebPQuality = cfg.WebPQuality
		}
	}

	// CLI flag overrides
	if outDir != "" {
		opts.Out = outDir
	}
	if opts.Input == "" {
		return opts, fmt.Errorf("input path is required (pass [input] or set \"input\" in iconkit.json)")
	}

	if opts.Padding < 0 || opts.Padding >= 0.5 {
		return opts, fmt.Errorf("--pad must be between 0.0 and 0.5 (exclusive), got %v", opts.Padding)
	}

	if bgColor != "" {
		c, err := processor.ParseHexColor(bgColor)
		if err != nil {
			return opts, err
		}
		opts.BgColor = c
	}

	// Resolve sizes: preset > -s > config sizes > default
	if presetName != "" {
		p, ok := preset.Get(presetName)
		if !ok {
			return opts, fmt.Errorf("unknown preset: %s (available: %s)",
				presetName, strings.Join(preset.Names(), ", "))
		}
		if sizes != "" {
			fmt.Fprintln(os.Stderr, "Note: -s is ignored when -p is specified")
		}
		opts.Sizes = p.Sizes
		opts.Preset = presetName
	} else if sizes != "" {
		parsed, err := parseSizes(sizes)
		if err != nil {
			return opts, err
		}
		opts.Sizes = parsed
	}

	hasPureProcessing := opts.Radius > 0 || opts.RadiusPercent > 0 || opts.Padding > 0 || opts.BgColor != nil || opts.Maskable
	if presetName == "" && len(opts.Sizes) == 0 && hasPureProcessing && !opts.Ico {
		opts.OriginalSizeOutput = true
	}

	return opts, nil
}

func parseSizes(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("invalid sizes format (use: 16,32,64)")
		}
		result = append(result, n)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("invalid sizes format (use: 16,32,64)")
	}
	return result, nil
}
