package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/Tendo33/iconkit/internal/config"
	"github.com/Tendo33/iconkit/internal/preset"
	"github.com/Tendo33/iconkit/internal/runner"
)

var version = "dev"

var (
	sizes      string
	radius     int
	presetName string
	outDir     string
	force      bool
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "iconkit <input>",
	Short: "Icon processing CLI — resize & round corners in one command",
	Long: `iconkit is a developer-friendly CLI tool for icon processing.

It takes a single image (or a directory of images) and outputs multiple sizes
with optional rounded corners.

Examples:
  iconkit icon.png
  iconkit icon.png -s 16,32,64,128
  iconkit icon.png -r 20 -s 16,32,64,128
  iconkit icon.png -p web
  iconkit icon.png -p ios
  iconkit icon.png -p android
  iconkit ./assets/ -p web
  iconkit icon.png -c iconkit.json
  iconkit icon.png -r 24 -s 16,32,64 -o ./dist`,
	Version: version,
	Args:    cobra.ExactArgs(1),
	RunE:    run,
}

func init() {
	rootCmd.Flags().StringVarP(&sizes, "sizes", "s", "", "output sizes, comma-separated (e.g. 16,32,64)")
	rootCmd.Flags().IntVarP(&radius, "radius", "r", 0, "corner radius in pixels")
	rootCmd.Flags().StringVarP(&presetName, "preset", "p", "", "size preset (web, ios, android)")
	rootCmd.Flags().StringVarP(&outDir, "out", "o", "", "output directory (default: ./icons)")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path (default: auto-detect iconkit.json)")
}

func Execute() error {
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	opts, err := buildOptions(inputPath)
	if err != nil {
		return err
	}

	results, err := runner.Run(opts, os.Stdout)
	if err != nil {
		return err
	}

	fmt.Printf("\nDone! %d icons saved to %s\n", len(results), opts.Out)
	return nil
}

func buildOptions(inputPath string) (runner.Options, error) {
	opts := runner.Options{
		Input:  inputPath,
		Radius: radius,
		Out:    "./icons",
		Force:  force,
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
		if cfg.Radius > 0 && radius == 0 {
			opts.Radius = cfg.Radius
		}
		if cfg.Out != "" && outDir == "" {
			opts.Out = cfg.Out
		}
		if cfg.Force && !force {
			opts.Force = cfg.Force
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
