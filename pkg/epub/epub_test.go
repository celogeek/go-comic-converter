package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celogeek/go-comic-converter/v3/pkg/epuboptions"
)

// writeNoisyImages creates n random images: they compress badly, so with LimitMb = 1 each one ends up in its own part.
func writeNoisyImages(t *testing.T, dir string, n int) {
	r := rand.New(rand.NewSource(1))
	for i := 1; i <= n; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 600, 900))
		r.Read(img.Pix)
		for p := 3; p < len(img.Pix); p += 4 {
			img.Pix[p] = 255
		}
		f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%02d.png", i)))
		require.NoError(t, err)
		require.NoError(t, png.Encode(f, img))
		require.NoError(t, f.Close())
	}
}

func testOptions(input, output string, coverCaption bool) epuboptions.EPUBOptions {
	return epuboptions.EPUBOptions{
		Input:        input,
		Output:       output,
		Title:        "My Comic",
		TitlePage:    1,
		CoverCaption: coverCaption,
		LimitMb:      1,
		SortPathMode: 1,
		Quiet:        true,
		Workers:      2,
		Image: epuboptions.Image{
			Quality: 90,
			Format:  "jpeg",
			Resize:  true,
			View: epuboptions.View{
				Width:        600,
				Height:       800,
				AspectRatio:  -1,
				PortraitOnly: true,
				Color:        epuboptions.Color{Foreground: "000000", Background: "FFFFFF"},
			},
		},
	}
}

func readZipFile(t *testing.T, z *zip.ReadCloser, name string) []byte {
	f, err := z.Open(name)
	require.NoError(t, err, name)
	defer func() { _ = f.Close() }()
	b, err := io.ReadAll(f)
	require.NoError(t, err, name)
	return b
}

func TestWriteSplitTitlePageAndCoverCaption(t *testing.T) {
	input := t.TempDir()
	writeNoisyImages(t, input, 3)

	output := t.TempDir()
	require.NoError(t, New(testOptions(input, filepath.Join(output, "comic.epub"), true)).Write())
	require.NoError(t, New(testOptions(input, filepath.Join(output, "nocaption.epub"), false)).Write())

	for _, label := range []string{"Part 1 of 3", "Part 2 of 3", "Part 3 of 3"} {
		t.Run(label, func(t *testing.T) {
			z, err := zip.OpenReader(filepath.Join(output, "comic - "+label+".epub"))
			require.NoError(t, err)
			defer func() { _ = z.Close() }()

			assert.Contains(t, string(readZipFile(t, z, "OEBPS/content.opf")), "My Comic - "+label)

			title, err := jpeg.Decode(bytes.NewReader(readZipFile(t, z, "OEBPS/Images/title.jpeg")))
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, 600, 800), title.Bounds(), "title page has the view size")
			// The title and the part label are on 2 lines centered on the page: look for dark pixels around the middle.
			darkRows := 0
			for y := 300; y < 500; y++ {
				for x := 0; x < 600; x++ {
					if color.GrayModel.Convert(title.At(x, y)).(color.Gray).Y < 128 {
						darkRows++
						break
					}
				}
			}
			assert.Greater(t, darkRows, 20, "title page has dark text in the middle")
			assert.Equal(t, color.Gray{Y: 255}, color.GrayModel.Convert(title.At(300, 10)), "title page background")

			noCaption, err := zip.OpenReader(filepath.Join(output, "nocaption - "+label+".epub"))
			require.NoError(t, err)
			defer func() { _ = noCaption.Close() }()
			assert.NotEqual(t,
				readZipFile(t, noCaption, "OEBPS/Images/cover.jpeg"),
				readZipFile(t, z, "OEBPS/Images/cover.jpeg"),
				"cover caption changes the cover",
			)
		})
	}
}
