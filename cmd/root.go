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
	sizes      string
	radius     int
	presetName string
	outDir     string
	force      bool
	configFile string
	padding    float64
	bgColor    string
	ico        bool
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
  iconkit icon.png -r 20 -s 16,32,64,128
  iconkit icon.png -p web
  iconkit icon.png -p chrome-ext
  iconkit icon.png -p firefox-ext
  iconkit icon.png --ico -p web
  iconkit icon.png --pad 0.1 --bg "#ffffff"
  iconkit -c iconkit.json
  iconkit icon.png -r 24 -s 16,32,64 -o ./dist`,
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	RunE:    run,
}

func init() {
	rootCmd.Flags().StringVarP(&sizes, "sizes", "s", "", "output sizes, comma-separated; overrides processing-only mode (e.g. 16,32,64)")
	rootCmd.Flags().IntVarP(&radius, "radius", "r", 0, "corner radius in pixels; without -s/-p outputs one PNG at the original size")
	rootCmd.Flags().StringVarP(&presetName, "preset", "p", "", "size preset (web, ios, android, chrome-ext, firefox-ext, pwa)")
	rootCmd.Flags().StringVarP(&outDir, "out", "o", "", "output directory (default: ./icons)")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path (default: auto-detect iconkit.json)")
	rootCmd.Flags().Float64Var(&padding, "pad", 0, "padding ratio around icon (0.0-0.5, e.g. 0.1 = 10%); without -s/-p outputs one PNG at the original size")
	rootCmd.Flags().StringVar(&bgColor, "bg", "", "background color in hex (e.g. \"#ffffff\", \"ff0000\"); without -s/-p outputs one PNG at the original size")
	rootCmd.Flags().BoolVar(&ico, "ico", false, "also generate favicon.ico (sizes <= 256); keeps multi-size output")
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
	if err != nil {
		return err
	}

	fmt.Printf("\nDone! %d files saved to %s\n", len(results), opts.Out)
	return nil
}

func buildOptions(inputPath string) (runner.Options, error) {
	opts := runner.Options{
		Input:   inputPath,
		Radius:  radius,
		Out:     "./icons",
		Force:   force,
		Padding: padding,
		Ico:     ico,
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
			fmt.Println("Note: -s is ignored when -p is specified")
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

	hasPureProcessing := opts.Radius > 0 || opts.Padding > 0 || opts.BgColor != nil
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
