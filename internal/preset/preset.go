package preset

import "sort"

type Preset struct {
	Sizes       []int
	Description string
}

var Registry = map[string]Preset{
	"web": {
		Sizes:       []int{16, 32, 48, 64, 128, 256},
		Description: "Web favicons & PWA icons",
	},
	"ios": {
		Sizes:       []int{20, 29, 40, 58, 60, 76, 80, 87, 120, 152, 167, 180, 1024},
		Description: "iOS AppIcon (all required sizes)",
	},
	"android": {
		Sizes:       []int{36, 48, 72, 96, 144, 192, 512},
		Description: "Android mipmap (mdpi → xxxhdpi + Play Store)",
	},
	"chrome-ext": {
		Sizes:       []int{16, 32, 48, 128},
		Description: "Chrome Extension (Manifest V3)",
	},
	"firefox-ext": {
		Sizes:       []int{32, 48, 64, 96, 128},
		Description: "Firefox Add-on",
	},
	"pwa": {
		Sizes:       []int{192, 512},
		Description: "Progressive Web App (minimum required)",
	},
	"macos": {
		Sizes:       []int{16, 32, 64, 128, 256, 512, 1024},
		Description: "macOS App Icon (all required sizes for .icns)",
	},
	"windows": {
		Sizes:       []int{16, 24, 32, 48, 64, 128, 256},
		Description: "Windows Shell + Microsoft Store icon sizes",
	},
	"electron": {
		Sizes:       []int{16, 32, 48, 64, 128, 256, 512, 1024},
		Description: "Electron cross-platform app icons",
	},
	"tauri": {
		Sizes:       []int{32, 128, 256},
		Description: "Tauri v2 app icons (use --output-name \"{width}x{height}\" for correct filenames)",
	},
}

func Get(name string) (Preset, bool) {
	p, ok := Registry[name]
	return p, ok
}

func Names() []string {
	names := make([]string, 0, len(Registry))
	for k := range Registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
