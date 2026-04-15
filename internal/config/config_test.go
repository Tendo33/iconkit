package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when file doesn't exist")
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	data := `{
		"input": "icon.png",
		"sizes": [16, 32, 64],
		"radius": 20,
		"preset": "web",
		"out": "./dist",
		"force": true
	}`
	os.WriteFile(filepath.Join(dir, DefaultFileName), []byte(data), 0o644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Input != "icon.png" {
		t.Errorf("input = %q, want %q", cfg.Input, "icon.png")
	}
	if len(cfg.Sizes) != 3 {
		t.Errorf("sizes len = %d, want 3", len(cfg.Sizes))
	}
	if cfg.Radius != 20 {
		t.Errorf("radius = %d, want 20", cfg.Radius)
	}
	if cfg.Preset != "web" {
		t.Errorf("preset = %q, want %q", cfg.Preset, "web")
	}
	if cfg.Out != "./dist" {
		t.Errorf("out = %q, want %q", cfg.Out, "./dist")
	}
	if !cfg.Force {
		t.Error("force should be true")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, DefaultFileName), []byte(`{broken`), 0o644)

	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoad_PartialJSON(t *testing.T) {
	dir := t.TempDir()
	data := `{"radius": 10}`
	os.WriteFile(filepath.Join(dir, DefaultFileName), []byte(data), 0o644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Radius != 10 {
		t.Errorf("radius = %d, want 10", cfg.Radius)
	}
	if cfg.Input != "" {
		t.Errorf("input should be empty, got %q", cfg.Input)
	}
	if len(cfg.Sizes) != 0 {
		t.Errorf("sizes should be empty, got %v", cfg.Sizes)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.json")
	data := `{"sizes": [128, 256], "out": "./out"}`
	os.WriteFile(path, []byte(data), 0o644)

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sizes) != 2 {
		t.Errorf("sizes len = %d, want 2", len(cfg.Sizes))
	}
	if cfg.Out != "./out" {
		t.Errorf("out = %q, want %q", cfg.Out, "./out")
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/iconkit.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
