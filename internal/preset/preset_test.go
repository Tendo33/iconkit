package preset

import (
	"testing"
)

func TestGet_Web(t *testing.T) {
	p, ok := Get("web")
	if !ok {
		t.Fatal("web preset should exist")
	}
	if len(p.Sizes) != 6 {
		t.Errorf("web preset should have 6 sizes, got %d", len(p.Sizes))
	}
	expected := []int{16, 32, 48, 64, 128, 256}
	for i, s := range expected {
		if p.Sizes[i] != s {
			t.Errorf("web sizes[%d] = %d, want %d", i, p.Sizes[i], s)
		}
	}
}

func TestGet_iOS(t *testing.T) {
	p, ok := Get("ios")
	if !ok {
		t.Fatal("ios preset should exist")
	}
	assertContains(t, "ios", p.Sizes, 1024) // App Store
}

func TestGet_Android(t *testing.T) {
	p, ok := Get("android")
	if !ok {
		t.Fatal("android preset should exist")
	}
	assertContains(t, "android", p.Sizes, 512) // Play Store
}

func TestGet_ChromeExt(t *testing.T) {
	p, ok := Get("chrome-ext")
	if !ok {
		t.Fatal("chrome-ext preset should exist")
	}
	for _, required := range []int{16, 32, 48, 128} {
		assertContains(t, "chrome-ext", p.Sizes, required)
	}
}

func TestGet_FirefoxExt(t *testing.T) {
	p, ok := Get("firefox-ext")
	if !ok {
		t.Fatal("firefox-ext preset should exist")
	}
	for _, required := range []int{32, 48, 64, 96, 128} {
		assertContains(t, "firefox-ext", p.Sizes, required)
	}
}

func TestGet_PWA(t *testing.T) {
	p, ok := Get("pwa")
	if !ok {
		t.Fatal("pwa preset should exist")
	}
	for _, required := range []int{192, 512} {
		assertContains(t, "pwa", p.Sizes, required)
	}
}

func TestGet_Unknown(t *testing.T) {
	_, ok := Get("nonexistent")
	if ok {
		t.Error("nonexistent preset should not be found")
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) < 6 {
		t.Errorf("expected at least 6 presets, got %d", len(names))
	}

	expected := map[string]bool{
		"web": false, "ios": false, "android": false,
		"chrome-ext": false, "firefox-ext": false, "pwa": false,
	}
	for _, n := range names {
		expected[n] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("preset %q not found in Names()", name)
		}
	}

	// Should be sorted
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("Names() should be sorted, got %v", names)
			break
		}
	}
}

func TestAllPresets_SizesPositive(t *testing.T) {
	for name, p := range Registry {
		if len(p.Sizes) == 0 {
			t.Errorf("preset %q has no sizes", name)
		}
		for _, s := range p.Sizes {
			if s <= 0 {
				t.Errorf("preset %q has invalid size %d", name, s)
			}
		}
		if p.Description == "" {
			t.Errorf("preset %q has no description", name)
		}
	}
}

func assertContains(t *testing.T, name string, sizes []int, target int) {
	t.Helper()
	for _, s := range sizes {
		if s == target {
			return
		}
	}
	t.Errorf("%s preset must include %d", name, target)
}
