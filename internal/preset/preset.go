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
