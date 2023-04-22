package epub

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/disintegration/gift"
	"github.com/gofrs/uuid"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomonobold"
)

type ImageOptions struct {
	Crop                bool
	ViewWidth           int
	ViewHeight          int
	Quality             int
	Algo                string
	Brightness          int
	Contrast            int
	AutoRotate          bool
	AutoSplitDoublePage bool
	NoBlankPage         bool
	Manga               bool
	HasCover            bool
	Workers             int
}

type EpubOptions struct {
	Input                      string
	Output                     string
	Title                      string
	Author                     string
	LimitMb                    int
	StripFirstDirectoryFromToc bool
	Dry                        bool
	DryVerbose                 bool
	SortPathMode               int
	Quiet                      bool

	*ImageOptions
}

type ePub struct {
	*EpubOptions
	UID       string
	Publisher string
	UpdatedAt string

	templateProcessor *template.Template
}

type epubPart struct {
	Cover  *Image
	Images []*Image
}

func NewEpub(options *EpubOptions) *ePub {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	tmpl := template.New("parser")
	tmpl.Funcs(template.FuncMap{
		"mod":  func(i, j int) bool { return i%j == 0 },
		"zoom": func(s int, z float32) int { return int(float32(s) * z) },
	})

	return &ePub{
		EpubOptions:       options,
		UID:               uid.String(),
		Publisher:         "GO Comic Converter",
		UpdatedAt:         time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		templateProcessor: tmpl,
	}
}

func (e *ePub) render(templateString string, data any) string {
	tmpl, err := e.templateProcessor.Parse(templateString)
	if err != nil {
		panic(err)
	}
	result := &strings.Builder{}
	if err := tmpl.Execute(result, data); err != nil {
		panic(err)
	}

	stripBlank := regexp.MustCompile("\n+")

	return stripBlank.ReplaceAllString(result.String(), "\n")
}

func (e *ePub) writeImage(wz *epubZip, img *Image) error {
	err := wz.WriteFile(
		fmt.Sprintf("OEBPS/%s", img.TextPath()),
		e.render(textTmpl, map[string]any{
			"Title":      fmt.Sprintf("Image %d Part %d", img.Id, img.Part),
			"ViewPort":   fmt.Sprintf("width=%d,height=%d", e.ViewWidth, e.ViewHeight),
			"ImagePath":  img.ImgPath(),
			"ImageStyle": img.ImgStyle(e.ViewWidth, e.ViewHeight, e.Manga),
		}),
	)

	if err == nil {
		err = wz.WriteImage(img.Data)
	}

	return err
}

func (e *ePub) writeBlank(wz *epubZip, img *Image) error {
	return wz.WriteFile(
		fmt.Sprintf("OEBPS/Text/%d_sp.xhtml", img.Id),
		e.render(blankTmpl, map[string]any{
			"Title":    fmt.Sprintf("Blank Page %d", img.Id),
			"ViewPort": fmt.Sprintf("width=%d,height=%d", e.ViewWidth, e.ViewHeight),
		}),
	)
}

func (e ePub) getTitleImageData(title string, img *Image, currentPart, totalPart int) *ImageData {
	// Create a blur version of the cover
	g := gift.New(gift.GaussianBlur(8))
	dst := image.NewGray(g.Bounds(img.Raw.Bounds()))
	g.Draw(dst, img.Raw)

	// Calculate size of title
	f, _ := truetype.Parse(gomonobold.TTF)
	borderSize := 4
	var fontSize, textWidth, textHeight int
	for fontSize = 64; fontSize >= 12; fontSize -= 1 {
		face := truetype.NewFace(f, &truetype.Options{Size: float64(fontSize), DPI: 72})
		textWidth = font.MeasureString(face, title).Ceil()
		textHeight = face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil()
		if textWidth+2*borderSize < img.Width && 3*textHeight+2*borderSize < img.Height {
			break
		}
	}

	// Draw rectangle in the middle of the image
	textPosStart := img.Height/2 - textHeight/2
	textPosEnd := img.Height/2 + textHeight/2
	marginSize := fontSize
	borderArea := image.Rect(0, textPosStart-borderSize-marginSize, img.Width, textPosEnd+borderSize+marginSize)
	textArea := image.Rect(borderSize, textPosStart-marginSize, img.Width-borderSize, textPosEnd+marginSize)

	draw.Draw(
		dst,
		borderArea,
		image.Black,
		image.Point{},
		draw.Over,
	)

	draw.Draw(
		dst,
		textArea,
		image.White,
		image.Point{},
		draw.Over,
	)

	// Draw text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFontSize(float64(fontSize))
	c.SetFont(f)
	c.SetClip(textArea)
	c.SetDst(dst)
	c.SetSrc(image.Black)

	textLeft := img.Width/2 - textWidth/2
	if textLeft < borderSize {
		textLeft = borderSize
	}
	c.DrawString(title, freetype.Pt(textLeft, img.Height/2+textHeight/4))

	return newData("OEBPS/Images/title.jpg", dst, e.Quality)
}

func (e *ePub) getParts() ([]*epubPart, error) {
	images, err := e.LoadImages()

	if err != nil {
		return nil, err
	}

	sort.Slice(images, func(i, j int) bool {
		if images[i].Id < images[j].Id {
			return true
		} else if images[i].Id == images[j].Id {
			return images[i].Part < images[j].Part
		} else {
			return false
		}
	})

	parts := make([]*epubPart, 0)
	cover := images[0]
	if e.HasCover {
		images = images[1:]
	}

	if e.Dry {
		parts = append(parts, &epubPart{
			Cover:  cover,
			Images: images,
		})
		return parts, nil
	}

	maxSize := uint64(e.LimitMb * 1024 * 1024)

	xhtmlSize := uint64(1024)
	// descriptor files + title
	baseSize := uint64(16*1024) + cover.Data.CompressedSize()
	if e.HasCover {
		baseSize += cover.Data.CompressedSize()
	}

	currentSize := baseSize
	currentImages := make([]*Image, 0)
	part := 1

	for _, img := range images {
		imgSize := img.Data.CompressedSize() + xhtmlSize
		if maxSize > 0 && len(currentImages) > 0 && currentSize+imgSize > maxSize {
			parts = append(parts, &epubPart{
				Cover:  cover,
				Images: currentImages,
			})
			part += 1
			currentSize = baseSize
			if !e.HasCover {
				currentSize += cover.Data.CompressedSize()
			}
			currentImages = make([]*Image, 0)
		}
		currentSize += imgSize
		currentImages = append(currentImages, img)
	}
	if len(currentImages) > 0 {
		parts = append(parts, &epubPart{
			Cover:  cover,
			Images: currentImages,
		})
	}

	return parts, nil
}

func (e *ePub) Write() error {
	type zipContent struct {
		Name    string
		Content any
	}

	epubParts, err := e.getParts()
	if err != nil {
		return err
	}

	if e.Dry {
		p := epubParts[0]
		fmt.Fprintf(os.Stderr, "TOC:\n  - %s\n%s\n", e.Title, e.getTree(p.Images, true))
		if e.DryVerbose {
			if e.HasCover {
				fmt.Fprintf(os.Stderr, "Cover:\n%s\n", e.getTree([]*Image{p.Cover}, false))
			}
			fmt.Fprintf(os.Stderr, "Files:\n%s\n", e.getTree(p.Images, false))
		}
		return nil
	}

	totalParts := len(epubParts)

	bar := NewBar(e.Quiet, totalParts, "Writing Part", 2, 2)
	for i, part := range epubParts {
		ext := filepath.Ext(e.Output)
		suffix := ""
		if totalParts > 1 {
			fmtLen := len(fmt.Sprint(totalParts))
			fmtPart := fmt.Sprintf(" Part %%0%dd of %%0%dd", fmtLen, fmtLen)
			suffix = fmt.Sprintf(fmtPart, i+1, totalParts)
		}

		path := fmt.Sprintf("%s%s%s", e.Output[0:len(e.Output)-len(ext)], suffix, ext)
		wz, err := newEpubZip(path)
		if err != nil {
			return err
		}
		defer wz.Close()

		title := e.Title
		if totalParts > 1 {
			title = fmt.Sprintf("%s [%d/%d]", title, i+1, totalParts)
		}

		content := []zipContent{
			{"META-INF/container.xml", containerTmpl},
			{"META-INF/com.apple.ibooks.display-options.xml", appleBooksTmpl},
			{"OEBPS/content.opf", e.getContent(title, part, i+1, totalParts).String()},
			{"OEBPS/toc.xhtml", e.getToc(title, part.Images)},
			{"OEBPS/Text/style.css", e.render(styleTmpl, map[string]any{
				"PageWidth":  e.ViewWidth,
				"PageHeight": e.ViewHeight,
			})},
			{"OEBPS/Text/title.xhtml", e.render(textTmpl, map[string]any{
				"Title":      title,
				"ViewPort":   fmt.Sprintf("width=%d,height=%d", e.ViewWidth, e.ViewHeight),
				"ImagePath":  "Images/title.jpg",
				"ImageStyle": part.Cover.ImgStyle(e.ViewWidth, e.ViewHeight, e.Manga),
			})},
		}

		if err = wz.WriteMagic(); err != nil {
			return err
		}
		for _, content := range content {
			if err := wz.WriteFile(content.Name, content.Content); err != nil {
				return err
			}
		}
		if err := wz.WriteImage(e.getTitleImageData(title, part.Cover, i+1, totalParts)); err != nil {
			return err
		}

		// Cover exist or part > 1
		// If no cover, part 2 and more will include the image as a cover
		if e.HasCover || i > 0 {
			if err := e.writeImage(wz, part.Cover); err != nil {
				return err
			}
		}

		for i, img := range part.Images {
			if err := e.writeImage(wz, img); err != nil {
				return err
			}

			// Double Page or Last Image
			if img.DoublePage || (i+1 == len(part.Images)) {
				if err := e.writeBlank(wz, img); err != nil {
					return err
				}
			}
		}
		bar.Add(1)
	}
	bar.Close()

	return nil
}
