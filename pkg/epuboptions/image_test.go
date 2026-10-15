package epuboptions

import "testing"

// a cut between 2 tiles is always at a multiple of the tile size: with a multiple of 16 (the size of a JPEG block),
// no tile ends with a partial block at a seam.
func TestDeviceTiles(t *testing.T) {
	for device, tile := range deviceTiles {
		if !tile.Limited() {
			t.Errorf("%s: tile %dx%d must be limited", device, tile.Width, tile.Height)
		}
		if tile.Width%16 != 0 || tile.Height%16 != 0 {
			t.Errorf("%s: tile %dx%d must be a multiple of 16", device, tile.Width, tile.Height)
		}
		if tile.Width*tile.Height >= 5_600_000 {
			t.Errorf("%s: tile %dx%d has more than 5.6 million pixels", device, tile.Width, tile.Height)
		}
	}
}

// only the devices with a constraint have a tile.
func TestImageTile(t *testing.T) {
	for _, c := range []struct {
		device  DeviceType
		limited bool
	}{
		{Kindle, true},
		{Apple, true},
		{Kobo, false},
		{ReMarkable, false},
		{"", false},
	} {
		if got := (Image{Device: c.device}).Tile().Limited(); got != c.limited {
			t.Errorf("%q: got limited %v, want %v", c.device, got, c.limited)
		}
	}
}

// the landscape only mode needs the option and a profile that has it.
func TestLandscapeMode(t *testing.T) {
	for _, c := range []struct {
		landscape, profile, want bool
	}{
		{false, false, false},
		{true, false, false},
		{false, true, false},
		{true, true, true},
	} {
		if got := (Image{Landscape: c.landscape, Features: Features{Landscape: c.profile}}).LandscapeMode(); got != c.want {
			t.Errorf("LandscapeMode(%v, %v) = %v, want %v", c.landscape, c.profile, got, c.want)
		}
	}
}

// the features of the profile win over the options: the converter can be called without the checks of the command line.
func TestMangaMode(t *testing.T) {
	for _, c := range []struct {
		manga, profile, want bool
	}{
		{false, false, false},
		{true, false, false},
		{false, true, false},
		{true, true, true},
	} {
		mode := ModeComic
		if c.manga {
			mode = ModeManga
		}
		if got := (Image{Mode: mode, Features: Features{Manga: c.profile}}).MangaMode(); got != c.want {
			t.Errorf("MangaMode(%v, %v) = %v, want %v", c.manga, c.profile, got, c.want)
		}
	}
}

func TestIsGrayScale(t *testing.T) {
	for _, c := range []struct {
		grayscale, color, want bool
	}{
		{false, false, true}, // no color screen: always gray
		{true, false, true},
		{false, true, false},
		{true, true, true},
	} {
		if got := (Image{GrayScale: c.grayscale, Features: Features{Color: c.color}}).IsGrayScale(); got != c.want {
			t.Errorf("IsGrayScale(%v, %v) = %v, want %v", c.grayscale, c.color, got, c.want)
		}
	}
}

// the landscape only mode doesn't support the rotation: the double page is split, whatever the mode.
func TestLandscapeModeSplitsDoublePage(t *testing.T) {
	for _, mode := range []int{DoublePageSplit, DoublePageSplitRotate, DoublePageRotate} {
		i := Image{DoublePageMode: mode, Landscape: true, Features: Features{Landscape: true}}
		if !i.SplitDoublePage() || i.RotateDoublePage() || i.KeepDoublePage() {
			t.Errorf("mode %d in landscape: split=%v rotate=%v keep=%v", mode, i.SplitDoublePage(), i.RotateDoublePage(), i.KeepDoublePage())
		}
	}

	// without the landscape mode the mode is the one of the option
	i := Image{DoublePageMode: DoublePageRotate}
	if i.SplitDoublePage() || !i.RotateDoublePage() || !i.KeepDoublePage() {
		t.Error("the rotate mode should be kept without the landscape mode")
	}
}

// the webtoon ignores the options about the double pages, the crop and the blank pages.
func TestWebtoonMode(t *testing.T) {
	img := Image{Crop: Crop{Enabled: true}, NoBlankImage: true, CombineSplitDoublePage: true}
	if !img.CropPages() || !img.RemoveBlankPages() || !img.CombineDoublePages() || img.WebtoonMode() {
		t.Errorf("the options must apply to a comic: %+v", img)
	}

	img.Mode = ModeWebtoon
	if img.CropPages() || img.RemoveBlankPages() || img.CombineDoublePages() || !img.WebtoonMode() {
		t.Errorf("the options must not apply to a webtoon: %+v", img)
	}
}

// the pages of a webtoon have the aspect ratio of the view when it is set, otherwise the one of the device.
func TestWebtoonAspectRatio(t *testing.T) {
	for _, c := range []struct {
		aspectRatio float64
		want        float64
	}{
		{-1, 2},
		{0, 2},
		{1.6, 1.6},
		{10, MaxAspectRatio},
	} {
		img := Image{View: View{Width: 100, Height: 200, AspectRatio: c.aspectRatio}}
		if got := img.WebtoonAspectRatio(); got != c.want {
			t.Errorf("WebtoonAspectRatio(%v) = %v, want %v", c.aspectRatio, got, c.want)
		}
	}
}
