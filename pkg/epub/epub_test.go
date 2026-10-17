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
	"strings"
	"testing"

	_ "github.com/gen2brain/jxl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celogeek/go-comic-converter/v3/internal/pkg/epubimage"
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
		Workers:      2,
		Image: epuboptions.Image{
			Quality: 90,
			Format:  "jpeg",
			Resize:  true,
			View: epuboptions.View{
				Width:       600,
				Height:      800,
				AspectRatio: -1,
				Color:       epuboptions.Color{Foreground: "000000", Background: "FFFFFF"},
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

// the cover and the title are in the output format, jpeg in the copy mode.
func TestWriteCoverTitleFormat(t *testing.T) {
	input := t.TempDir()
	writeNoisyImages(t, input, 2)

	for format, want := range map[string]string{"jpeg": "jpeg", "png": "png", "jxl": "jxl", "copy": "jpeg"} {
		t.Run(format, func(t *testing.T) {
			o := testOptions(input, filepath.Join(t.TempDir(), "comic.epub"), true)
			o.LimitMb = 0
			o.Image.Format = format
			require.NoError(t, New(o).Write())

			z, err := zip.OpenReader(o.Output)
			require.NoError(t, err)
			defer func() { _ = z.Close() }()

			for _, name := range []string{"cover", "title"} {
				img, decoded, err := image.Decode(bytes.NewReader(readZipFile(t, z, "OEBPS/Images/"+name+"."+want)))
				require.NoError(t, err, name)
				assert.Equal(t, want, decoded, name)
				assert.False(t, img.Bounds().Empty(), name)
			}
			opf := string(readZipFile(t, z, "OEBPS/content.opf"))
			assert.Contains(t, opf, `href="Images/cover.`+want+`" id="img_cover" media-type="image/`+want+`"`)
			assert.Contains(t, opf, `href="Images/title.`+want+`" id="img_title" media-type="image/`+want+`"`)
			assert.Contains(t, string(readZipFile(t, z, "OEBPS/Text/title.xhtml")), `src="../Images/title.`+want+`"`)
		})
	}
}

// the temporary image storage must be removed whatever the outcome.
func TestWriteRemovesTemporaryStorage(t *testing.T) {
	input := t.TempDir()
	writeNoisyImages(t, input, 2)
	bad := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bad, "bad.jpg"), []byte("bad"), 0o644))
	empty := t.TempDir()

	for _, c := range []struct {
		name    string
		input   string
		format  string
		dry     bool
		wantErr bool
	}{
		{"convert", input, "jpeg", false, false},
		{"convert dry", input, "jpeg", true, false},
		{"copy dry", input, "copy", true, false},
		{"copy unreadable image", bad, "copy", false, false},
		{"no images", empty, "jpeg", false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "comic.epub")
			o := testOptions(c.input, output, false)
			o.Image.Format = c.format
			o.Dry = c.dry

			err := New(o).Write()
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			_, err = os.Stat(o.Output + ".tmp")
			assert.True(t, os.IsNotExist(err), "temporary storage left behind: %v", err)
		})
	}
}

// the copy mode never splits: a double page at the start of the book is a page, not only the cover.
func TestWriteCopyKeepsLeadingDoublePage(t *testing.T) {
	input := t.TempDir()
	for name, size := range map[string]image.Rectangle{
		"001.jpg": image.Rect(0, 0, 400, 200),
		"002.jpg": image.Rect(0, 0, 200, 400),
	} {
		f, err := os.Create(filepath.Join(input, name))
		require.NoError(t, err)
		require.NoError(t, jpeg.Encode(f, image.NewGray(size), nil))
		require.NoError(t, f.Close())
	}

	o := testOptions(input, filepath.Join(t.TempDir(), "comic.epub"), false)
	o.Image.Format = "copy"
	o.TitlePage = 0
	require.NoError(t, New(o).Write())

	z, err := zip.OpenReader(o.Output)
	require.NoError(t, err)
	defer func() { _ = z.Close() }()

	pages := 0
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "OEBPS/Text/page_") {
			pages++
		}
	}
	assert.Equal(t, 2, pages, "both images are pages")
}

// the aspect ratio of the source can be extreme, the view is never thinner than the limits.
func TestComputeViewPortLimitsSourceAspectRatio(t *testing.T) {
	for _, c := range []struct {
		name  string
		ratio float64
		want  [2]int
	}{
		{"extreme tall", 1000, [2]int{266, 800}},
		{"extreme wide", 0.0001, [2]int{600, 300}},
		{"in the limits", 1.6, [2]int{500, 800}},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := New(testOptions("", "", false)).(epub)
			e.Image.View.AspectRatio = 0
			parts := []epubPart{{Cover: epubimage.EPUBImage{OriginalAspectRatio: c.ratio}}}

			w, h := e.computeViewPort(parts)
			assert.Equal(t, c.want, [2]int{w, h})
		})
	}
}
