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
