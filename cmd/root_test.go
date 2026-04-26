package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSizes_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"16,32,64", []int{16, 32, 64}},
		{"128", []int{128}},
		{"16, 32, 64", []int{16, 32, 64}},
		{" 16 , 32 ", []int{16, 32}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSizes(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("len = %d, want %d", len(result), len(tt.expected))
			}
			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("result[%d] = %d, want %d", i, result[i], v)
				}
			}
		})
	}
}

func TestParseSizes_Invalid(t *testing.T) {
	invalid := []string{
		"abc",
		"16,abc,32",
		"",
		",,,",
		"0",
		"-1",
		"16,-5",
	}

	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			_, err := parseSizes(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestBuildOptions_UsesExtendedConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configJSON := `{
		"pad": 0.1,
		"bg": "#112233",
		"ico": true
	}`
	if err := os.WriteFile(filepath.Join(tempDir, "iconkit.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
		resetRootFlags()
	})

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	resetRootFlags()

	opts, err := buildOptions("icon.png")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}

	if opts.Padding != 0.1 {
		t.Fatalf("padding = %v, want 0.1", opts.Padding)
	}
	if !opts.Ico {
		t.Fatal("ico = false, want true")
	}
	if opts.BgColor == nil {
		t.Fatal("bgColor = nil, want parsed color")
	}

	r, g, b, a := opts.BgColor.RGBA()
	if r>>8 != 0x11 || g>>8 != 0x22 || b>>8 != 0x33 || a>>8 != 0xff {
		t.Fatalf("bgColor = %#02x %#02x %#02x %#02x, want 11 22 33 ff", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestBuildOptions_UsesConfigInputWhenArgMissing(t *testing.T) {
	tempDir := t.TempDir()
	configJSON := `{
		"input": "icon.png"
	}`
	if err := os.WriteFile(filepath.Join(tempDir, "iconkit.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
		resetRootFlags()
	})

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	resetRootFlags()

	opts, err := buildOptions("")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}
	if opts.Input != "icon.png" {
		t.Fatalf("input = %q, want %q", opts.Input, "icon.png")
	}
}

func TestRootCmd_AllowsOptionalInputArgForConfigInput(t *testing.T) {
	if err := rootCmd.Args(rootCmd, []string{}); err != nil {
		t.Fatalf("expected zero args to be allowed, got %v", err)
	}
}

func TestRootCmd_HelpTextHasNoMojibake(t *testing.T) {
	checks := []string{
		rootCmd.Short,
		rootCmd.Long,
		rootCmd.Flag("sizes").Usage,
		rootCmd.Flag("radius").Usage,
		rootCmd.Flag("pad").Usage,
		rootCmd.Flag("bg").Usage,
		rootCmd.Flag("ico").Usage,
	}

	for _, text := range checks {
		if strings.Contains(text, "鈥") || strings.Contains(text, "鈮") || strings.Contains(text, "�") {
			t.Fatalf("unexpected mojibake in help text: %q", text)
		}
	}

	if got := rootCmd.Flag("sizes").Usage; got != "output sizes, comma-separated; overrides processing-only mode (e.g. 16,32,64)" {
		t.Fatalf("sizes usage = %q", got)
	}
	if got := rootCmd.Flag("radius").Usage; got != "corner radius in pixels; without -s/-p outputs one PNG at the original size" {
		t.Fatalf("radius usage = %q", got)
	}
	if got := rootCmd.Flag("pad").Usage; got != "padding ratio around icon (0.0-0.5, e.g. 0.1 = 10%); without -s/-p outputs one PNG at the original size" {
		t.Fatalf("pad usage = %q", got)
	}
	if got := rootCmd.Flag("bg").Usage; got != "background color in hex (e.g. \"#ffffff\", \"ff0000\"); without -s/-p outputs one PNG at the original size" {
		t.Fatalf("bg usage = %q", got)
	}
	if got := rootCmd.Flag("ico").Usage; got != "also generate favicon.ico (sizes <= 256); keeps multi-size output" {
		t.Fatalf("ico usage = %q", got)
	}
}

func TestRootCmd_HelpText_DescribesProcessingOnlyMode(t *testing.T) {
	if !strings.Contains(rootCmd.Long, "Without -s or -p, using -r, --pad, or --bg keeps the original dimensions") {
		t.Fatalf("rootCmd.Long should describe processing-only mode, got:\n%s", rootCmd.Long)
	}

	if !strings.Contains(rootCmd.Long, "In batch mode, files are written as {name}.png") {
		t.Fatalf("rootCmd.Long should describe batch naming, got:\n%s", rootCmd.Long)
	}

	if !strings.Contains(rootCmd.Long, "same-name conflicts become {name}-{source-ext}.png") {
		t.Fatalf("rootCmd.Long should describe collision naming, got:\n%s", rootCmd.Long)
	}

	if !strings.Contains(rootCmd.Flag("ico").Usage, "keeps multi-size output") {
		t.Fatalf("ico usage should mention multi-size output, got %q", rootCmd.Flag("ico").Usage)
	}
}

func TestBuildOptions_RejectsPaddingOutOfRange(t *testing.T) {
	resetRootFlags()
	t.Cleanup(resetRootFlags)

	padding = 0.5
	_, err := buildOptions("icon.png")
	if err == nil {
		t.Fatal("expected error for padding = 0.5")
	}

	padding = -0.1
	_, err = buildOptions("icon.png")
	if err == nil {
		t.Fatal("expected error for negative padding")
	}

	padding = 0.6
	_, err = buildOptions("icon.png")
	if err == nil {
		t.Fatal("expected error for padding = 0.6")
	}
}

func TestBuildOptions_RadiusWithoutSizesUsesOriginalSizeOutput(t *testing.T) {
	resetRootFlags()
	t.Cleanup(resetRootFlags)

	radius = 24

	opts, err := buildOptions("icon.png")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}

	if !opts.OriginalSizeOutput {
		t.Fatal("OriginalSizeOutput = false, want true")
	}
}

func TestBuildOptions_RadiusWithSizesKeepsMultiSizeOutput(t *testing.T) {
	resetRootFlags()
	t.Cleanup(resetRootFlags)

	radius = 24
	sizes = "16,32"

	opts, err := buildOptions("icon.png")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}

	if opts.OriginalSizeOutput {
		t.Fatal("OriginalSizeOutput = true, want false")
	}
}

func TestBuildOptions_PaddingWithoutSizesUsesOriginalSizeOutput(t *testing.T) {
	resetRootFlags()
	t.Cleanup(resetRootFlags)

	padding = 0.1

	opts, err := buildOptions("icon.png")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}

	if !opts.OriginalSizeOutput {
		t.Fatal("OriginalSizeOutput = false, want true")
	}
}

func TestBuildOptions_BackgroundWithoutSizesUsesOriginalSizeOutput(t *testing.T) {
	resetRootFlags()
	t.Cleanup(resetRootFlags)

	bgColor = "#ffffff"

	opts, err := buildOptions("icon.png")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}

	if !opts.OriginalSizeOutput {
		t.Fatal("OriginalSizeOutput = false, want true")
	}
}

func TestBuildOptions_PureProcessingWithIcoKeepsMultiSizeOutput(t *testing.T) {
	resetRootFlags()
	t.Cleanup(resetRootFlags)

	padding = 0.1
	ico = true

	opts, err := buildOptions("icon.png")
	if err != nil {
		t.Fatalf("buildOptions: %v", err)
	}

	if opts.OriginalSizeOutput {
		t.Fatal("OriginalSizeOutput = true, want false when ico is enabled")
	}
}

func resetRootFlags() {
	sizes = ""
	radius = 0
	radiusPercent = 0
	presetName = ""
	outDir = ""
	force = false
	configFile = ""
	padding = 0
	bgColor = ""
	ico = false
	resizeMode = "stretch"
	outputName = ""
	generateHTML = false
	generateManifest = false
	maskable = false
	dryRun = false
	quiet = false
	verbose = false
	continueOnErr = false
	concurrency = 0
	format = "png"
	webpQuality = 90
}

func findPOSIXShell(t *testing.T) string {
	t.Helper()

	candidates := []string{
		"sh",
		"sh.exe",
		`C:\Program Files\Git\usr\bin\sh.exe`,
	}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	t.Skip("POSIX shell not available")
	return ""
}

func runInstallScriptForTest(t *testing.T, body string, extraEnv ...string) string {
	t.Helper()

	shellPath := findPOSIXShell(t)
	scriptPath, err := filepath.Abs(filepath.Join("..", "install.sh"))
	if err != nil {
		t.Fatalf("resolve install.sh: %v", err)
	}

	command := `. "$1"` + "\n" + body
	cmd := exec.Command(shellPath, "-c", command, "sh", scriptPath)
	cmd.Env = append(os.Environ(), append([]string{
		"ICONKIT_INSTALL_TEST_MODE=1",
		"SHELL=/usr/bin/zsh",
		"PATH=" + filepath.Dir(shellPath) + string(os.PathListSeparator) + os.Getenv("PATH"),
	}, extraEnv...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh command failed: %v\n%s", err, output)
	}

	return string(output)
}

func TestInstallScript_PostInstallMessageHintsRehashOnWindows(t *testing.T) {
	output := runInstallScriptForTest(t, `LATEST=v0.0.2
print_success_message "windows" "/usr/local/bin" "iconkit.exe"`)

	if !strings.Contains(output, "rehash") {
		t.Fatalf("expected Windows post-install message to mention rehash, got:\n%s", output)
	}

	if !strings.Contains(output, "/usr/local/bin/iconkit.exe") {
		t.Fatalf("expected installed binary path in output, got:\n%s", output)
	}
}

func TestInstallScript_CleanupCanRemoveTempDirFromInside(t *testing.T) {
	baseDir := t.TempDir()
	tmpDir := filepath.Join(baseDir, "tmp")
	restoreDir := filepath.Join(baseDir, "restore")

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp: %v", err)
	}
	if err := os.MkdirAll(restoreDir, 0o755); err != nil {
		t.Fatalf("mkdir restore: %v", err)
	}

	runInstallScriptForTest(t, `TMPDIR="$TMPDIR_ARG"
ORIG_PWD="$ORIG_PWD_ARG"
cd "$TMPDIR"
cleanup
test ! -d "$TMPDIR"`, "TMPDIR_ARG="+tmpDir, "ORIG_PWD_ARG="+restoreDir)
}
