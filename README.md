# go-comic-converter

Convert CBZ/CBR/Dir into EPUB for e-reader devices (Kindle Devices, ...)

My goal is to make a simple, cross-platform, and fast tool to convert comics into EPUB.

EPUB is now support by Amazon through [SendToKindle](https://www.amazon.com/gp/sendtokindle/), by Email or by using the App. So I've made it simple to support the size limit constraint of those services.

# Features
- Support input from zip, cbz, rar, cbr, pdf, directory
- Support all Kindle, Kobo and reMarkable devices
- Support Landscape and Portrait mode
- Customize output image quality
- Intelligent cropping (support removing even page numbers)
- Customize brightness and contrast
- Auto contrast
- Double page handling (never kept untouched): split, split + rotate, or rotate only
- Images bigger than the SR resolution (1200x1920), like with the KS or HR profile, are stored as a grid of tiles shown as one page, to keep the resolution on large devices
- Optionally detect a double page split in 2 consecutive pages, and combine them (`-combinesplitdoublepage`, off by default)
- Remove blank image (empty image is removed)
- Manga or Normal mode
- Embedded panel view for the older devices that need it (set by the profile)
- The cover is the first image that is not blank
- An unreadable image is replaced by a page with its name, instead of stopping the conversion
- Support title page (cover with embedded title and part)
- Split EPUB size for easy upload
- 3 sorting methods (depending on your source, you can ensure the page go in the right order)
- Save and reuse your own perfect settings
- Multi tasks for fast conversion
- Apple Book Compatibility Mode
- JSON output for programmatic usage
- Support cover caption with the part number when the EPUB is split
- Support for own font for the title page and the cover caption

When you read the comic on a Kindle, you can customize how you read it with the `Aa` button:
- Landscape / Portrait
- Activate panel view for small device

The panel view is handled natively by recent Kindles, for example the Kindle Scribe.
The embedded panel view comes only from the profile: it is enabled for the older devices that need it, and can't be changed from the command line or the saved config.

The grayscale can only be turned off (`-grayscale=0`) with a profile of a color device (`HR`, `SR`, `KCS`, `KSC`, `KoCC`, `KoLC`, `RMP`, `RMPM`).

# Installation

First ensure to have a working version of GO: [Installation](https://go.dev/doc/install)

Then install the last version of the tool:
```
$ go install github.com/celogeek/go-comic-converter/v3
```

To force install a specific version:
```
# specific version
$ go install github.com/celogeek/go-comic-converter/v3@v3.0.0

# main branch
$ go install github.com/celogeek/go-comic-converter/v3@main

# specific commit
$ go install github.com/celogeek/go-comic-converter/v3@COMMIT_HASH
```

Add GOPATH to your PATH
```
$ export PATH=$(go env GOPATH)/bin:$PATH
```

# Installation on Alpine

For Alpine user, you will need extra steps:
```
# install the external library
apk add mupdf-dev

# use it when you build it
go install -tags extlib github.com/celogeek/go-comic-converter/v3
```
If you prefer a pure go version:

```
# use it when you build it
go install -tags nocgo github.com/celogeek/go-comic-converter/v3
```

# Upgrade from V2

The configuration file structure changes in the v3 compare to v2.

You need to recreate your config and save it again.

Use the `show`, `reset` and `save` option.

# Check last version

You can check if a new version is available with:
```
$ go-comic-converter -version
go-comic-converter
  Path             : github.com/celogeek/go-comic-converter/v3
  Sum              : h1:tUFF2m/fGlOJOwC0/PlTopMfcBMprKvgr6TiQHQxEeo=
  Version          : v3.0.0
  Available Version: v3.0.0

To install the latest version:
$ go install github.com/celogeek/go-comic-converter/v3@v3.0.0
```

# Supported image files

The supported image files are jpeg and png from the sources.

The extensions can be: `jpg`, `jpeg`, `png`, `webp`, `tiff`.

The case for extensions doesn't matter.

For the passthrough mode (format=copy), the supported extensions are: `jpg`, `jpeg`, `png`

# Usage

## Convert directory

Convert every supported image files found in the input directory:

```
$ go-comic-converter -profile SR -input ~/Download/MyComic
```

By default, it will output: ~/Download/MyComic.epub

## Convert CBZ, ZIP, CBR, RAR, PDF

Convert every supported image files found in the input directory:

```
$ go-comic-converter -profile SR -input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF]
```

By default, it will output: ~/Download/MyComic.epub

## Convert with size limit

If you send your ePub through Amazon service, you have some size limitation:
  - Email  : 50Mb (including encoding, so 40Mb for RAW file)
  - App    : 50Mb
  - Website: 200Mb

By default, the EPUB is split every 200Mb (SendToKindle website limit).

You can change it using the "-limitmb MB" option, or disable it with "-limitmb 0":

```
go-comic-converter -profile SR -input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF] -limitmb 50
```

If you have more than 1 file the output will be:
  - ~/Download/MyComic - Part 01 of 10.epub
  - ~/Download/MyComic - Part 02 of 10.epub
  - ...

NUM is zero-padded to the number of digits of TOTAL (Part 01 of 10), in the file name, the title page, the EPUB title and the cover caption, so the parts sort in order.

The ePub include as a first page (see `-titlepage`):
  - Title
  - Part NUM of TOTAL

If the total is above 1:
  - the title of the EPUB becomes: Title - Part NUM of TOTAL
  - the cover shows a "Part NUM of TOTAL" caption (disable it with `-covercaption=0`)

The cover caption and the title page use the embedded [Literata](https://github.com/googlefonts/literata) font (SIL Open Font License). You can use your own TrueType font with `-font /path/to/font.ttf` (save it as default with `-save`).

## Dry run

If you want to preview what will be set during the conversion without running the conversion, then you can use the `-dry` option.

```
$ go-comic-converter -input ~/Downloads/mymanga.cbr -profile SR -manga -limitmb 200 -dry
Go Comic Converter

Options:
    Input                           : ~/Downloads/mymanga.cbr
    Output                          : ~/Downloads/mymanga.epub
    Author                          : GO Comic Converter
    Title                           : mymanga
    Workers                         : numcpus
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : 1
    Grayscale mode                  : normal
    Crop                            : 1
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip disabled
    Double page mode                : split
    Keep split double page aspect   : 1
    No blank image                  : 1
    Manga                           : 1
    Limit                           : 200 Mb
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000000
    Background color                : #FFFFFF
    Resize                          : 1
    Aspect ratio                    : auto
    Title page                      : always
    Cover caption                   : 1
    Font                            : Literata (embedded)

TOC:
  - mymanga
  - Chapter 1
  - Chapter 2
  - Chapter 3

Cover:
  - Chapter 1
    - img1.jpg

Files:
  - Chapter 1
    - img1.jpg
    - img10.jpg
    - img2.jpg
  - Chapter 2
    - img01.jpg
    - img02.jpg
    - img03.jpg
  - Chapter 3
    - img1.jpg
    - img2-3.jpg
    - img4.jpg
```

You can choose different way to sort path and files, depending on your source.

The option `sort` allow you to change the sorting order.

```
$ go-comic-converter -input ~/Downloads/mymanga.cbr -profile SR -manga -limitmb 200 -dry -sort 2
Go Comic Converter

Options:
    Input                           : ~/Downloads/mymanga.cbr
    Output                          : ~/Downloads/mymanga.epub
    Author                          : GO Comic Converter
...
    Sort path mode                  : path=alphanumeric, file=alphanumeric
...

TOC:
  - mymanga
  - Chapter 1
  - Chapter 2
  - Chapter 3

Cover:
  - Chapter 1
    - img1.jpg

Files:
  - Chapter 1
    - img1.jpg
    - img2.jpg
    - img10.jpg
  - Chapter 2
    - img01.jpg
    - img02.jpg
    - img03.jpg
  - Chapter 3
    - img1.jpg
    - img2-3.jpg
    - img4.jpg
```

## Change default settings

### Show current default option
```
$ go-comic-converter -show

Go Comic Converter

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : 1
    Grayscale mode                  : normal
    Crop                            : 1
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip disabled
    Double page mode                : split
    Keep split double page aspect   : 1
    No blank image                  : 1
    Limit                           : 200 Mb
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000000
    Background color                : #FFFFFF
    Resize                          : 1
    Aspect ratio                    : auto
    Title page                      : always
    Cover caption                   : 1
    Font                            : Literata (embedded)
```

### Change default settings
```
$ go-comic-converter -manga -profile SR -limitmb 200 -save

Go Comic Converter

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : 1
    Grayscale mode                  : normal
    Crop                            : 1
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip disabled
    Double page mode                : split
    Keep split double page aspect   : 1
    No blank image                  : 1
    Manga                           : 1
    Limit                           : 200 Mb
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000000
    Background color                : #FFFFFF
    Resize                          : 1
    Aspect ratio                    : auto
    Title page                      : always
    Cover caption                   : 1
    Font                            : Literata (embedded)

Saving to ~/.go-comic-converter.yaml
```

If you want to change a setting, you can change only one of them
```
$ go-comic-converter -manga=0 -save

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : 1
    Grayscale mode                  : normal
    Crop                            : 1
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip disabled
    Double page mode                : split
    Keep split double page aspect   : 1
    No blank image                  : 1
    Limit                           : 200 Mb
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000000
    Background color                : #FFFFFF
    Resize                          : 1
    Aspect ratio                    : auto
    Title page                      : always
    Cover caption                   : 1
    Font                            : Literata (embedded)

Saving to ~/.go-comic-converter.yaml
```

### Reset default
To reset all value to default:

```
$ go-comic-converter -reset
Go Comic Converter

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : 1
    Grayscale mode                  : normal
    Crop                            : 1
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip disabled
    Double page mode                : split
    Keep split double page aspect   : 1
    No blank image                  : 1
    Limit                           : 200 Mb
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000000
    Background color                : #FFFFFF
    Resize                          : 1
    Aspect ratio                    : auto
    Title page                      : always
    Cover caption                   : 1
    Font                            : Literata (embedded)

Reset default to ~/.go-comic-converter.yaml
```

# My own settings

After playing around with the options, I have my perfect settings for all my devices.

```
$ go-comic-converter -reset
$ go-comic-converter -profile KS -quality 90 -manga -aspect-ratio 1.6 -limitmb 200 -crop=0 -grayscale=1 -grayscale-mode=2 -titlepage=1 -covercaption=1 -combinesplitdoublepage=1 --doublepagemode=0 -autocontrast=1 -save

Options:
    Profile                         : KS - Kindle Scribe - 1860x2480
    Format                          : jpeg
    Quality                         : 90
    Grayscale                       : 1
    Grayscale mode                  : luminance
    Double page mode                : split
    Combine split double page       : 1
    Double page mode                : split
    Keep split double page aspect   : 1
    No blank image                  : 1
    Manga                           : 1
    Limit                           : 200 Mb
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000000
    Background color                : #FFFFFF
    Resize                          : 1
    Aspect ratio                    : 1:1.60
    Title page                      : always
    Cover caption                   : 1
    Font                            : Literata (embedded)

Saving to ~/.go-comic-converter.yaml
```

Explanation:
- `-profile KS`: my big device. put yours for best result.
- `-quality 90`: JPEG output quality of images
- `-manga`: manga mode, read right to left
- `-aspect-ratio`: ensure aspect ratio is 1:1.6, best for kindle devices.
- `-limitmb 200`: size limit to 200MB allowing upload from SendToKindle website
- `-crop=0`: do not remove anything
- `-grayscale=1 -grayscale-mode=2`: use luminance mode of grayscale color
- `-titlepage=1`: create a page with the title
- `-covercaption=1`: add a footer on epub that is split on the thumbnail
- `-combinesplitdoublepage=1`: combine pages that should be as 1 big picture
- `-doublepagemode=0`: split the double page, the kindle will show an option to see them entirely in landscape
- `-autocontrast=1`: improve automatically the contrast

# Help

```
$ go-comic-converter -h

Usage of go-comic-converter:

Output:
  -input=string
    	Source of comic to convert: directory, cbz, zip, cbr, rar, pdf
  -output=string
    	Output of the EPUB (directory or EPUB): (default [INPUT].epub)
  -author=string (default "GO Comic Converter")
    	Author of the EPUB
  -title=string
    	Title of the EPUB

Config:
  -profile=string (default "SR")
    	Profile to use: 
    	    Generic:
    	      - HR              - 2400 x 3840 - High Resolution
    	      - SR              - 1200 x 1920 - Standard Resolution
    	    Kindle:
    	      - K1       (2007) -  600 x 800  - Kindle 1
    	      - K2       (2009) -  600 x 800  - Kindle 2
    	      - KDX      (2009) -  824 x 1200 - Kindle DX/DXG
    	      - K3       (2010) -  600 x 800  - Kindle Keyboard
    	      - K4       (2011) -  600 x 800  - Kindle 4
    	      - KT       (2011) -  600 x 800  - Kindle Touch
    	      - K5       (2012) -  600 x 800  - Kindle 5
    	      - KPW1     (2012) -  758 x 1024 - Kindle Paperwhite 1
    	      - KPW2     (2013) -  758 x 1024 - Kindle Paperwhite 2
    	      - K7       (2014) -  600 x 800  - Kindle 7
    	      - KV       (2014) - 1072 x 1448 - Kindle Voyage
    	      - KPW3     (2015) - 1072 x 1448 - Kindle Paperwhite 3
    	      - KO1      (2016) - 1072 x 1448 - Kindle Oasis 1
    	      - K8       (2016) -  600 x 800  - Kindle 8
    	      - KO2      (2017) - 1264 x 1680 - Kindle Oasis 2
    	      - KPW4     (2018) - 1072 x 1448 - Kindle Paperwhite 4
    	      - K10      (2019) -  600 x 800  - Kindle 10
    	      - KO3      (2019) - 1264 x 1680 - Kindle Oasis 3
    	      - KPW5     (2021) - 1236 x 1648 - Kindle Paperwhite 5/Signature Edition
    	      - K11      (2022) - 1072 x 1448 - Kindle 11
    	      - KS       (2022) - 1860 x 2480 - Kindle Scribe
    	      - K12      (2024) - 1072 x 1448 - Kindle 12
    	      - KPW12    (2024) - 1264 x 1680 - Kindle Paperwhite 12/Signature Edition
    	      - KCS      (2024) - 1264 x 1680 - Kindle Colorsoft
    	      - KS2      (2024) - 1860 x 2480 - Kindle Scribe 2
    	      - KS3      (2025) - 1980 x 2640 - Kindle Scribe 3
    	      - KSC      (2025) - 1980 x 2640 - Kindle Scribe Colorsoft
    	    Kobo:
    	      - KoR      (2010) -  600 x 800  - Kobo eReader
    	      - KoT      (2011) -  600 x 800  - Kobo Touch
    	      - KoM      (2012) -  600 x 800  - Kobo Mini
    	      - KoG      (2012) -  758 x 1024 - Kobo Glo
    	      - KoA      (2013) -  758 x 1024 - Kobo Aura
    	      - KoAHD    (2013) - 1080 x 1440 - Kobo Aura HD
    	      - KoAH2O   (2014) - 1080 x 1430 - Kobo Aura H2O
    	      - KoGHD    (2015) - 1072 x 1448 - Kobo Glo HD
    	      - KoT2     (2015) -  600 x 800  - Kobo Touch 2.0
    	      - KoAO     (2016) - 1404 x 1872 - Kobo Aura One
    	      - KoA2     (2016) -  758 x 1024 - Kobo Aura Edition 2
    	      - KoAH2O2  (2017) - 1080 x 1430 - Kobo Aura H2O Edition 2
    	      - KoCHD    (2018) - 1072 x 1448 - Kobo Clara HD
    	      - KoF      (2018) - 1440 x 1920 - Kobo Forma
    	      - KoLH2O   (2019) - 1264 x 1680 - Kobo Libra H2O
    	      - KoN      (2020) -  758 x 1024 - Kobo Nia
    	      - KoE      (2021) - 1404 x 1872 - Kobo Elipsa
    	      - KoS      (2021) - 1440 x 1920 - Kobo Sage
    	      - KoL2     (2021) - 1264 x 1680 - Kobo Libra 2
    	      - KoC2E    (2022) - 1072 x 1448 - Kobo Clara 2E
    	      - KoE2     (2023) - 1404 x 1872 - Kobo Elipsa 2E
    	      - KoCBW    (2024) - 1072 x 1448 - Kobo Clara BW
    	      - KoCC     (2024) - 1072 x 1448 - Kobo Clara Colour
    	      - KoLC     (2024) - 1264 x 1680 - Kobo Libra Colour
    	    reMarkable:
    	      - RM1      (2017) - 1404 x 1872 - reMarkable 1
    	      - RM2      (2020) - 1404 x 1872 - reMarkable 2
    	      - RMP      (2024) - 1620 x 2160 - reMarkable Paper Pro
    	      - RMPM     (2025) -  954 x 1696 - reMarkable Paper Pro Move
    	      - RMPU     (2026) - 1404 x 1872 - reMarkable Paper Pure
  -quality=int (default 85)
    	Quality of the image
  -grayscale or -grayscale=1 (to enabled), -grayscale=0 (to disabled) (default 1)
    	Grayscale image. Ideal for eInk devices.
  -grayscale-mode=int
    	Grayscale Mode
    	0 = normal
    	1 = average
    	2 = luminance
  -crop or -crop=1 (to enabled), -crop=0 (to disabled) (default 1)
    	Crop images
  -crop-ratio-left=int (default 1)
    	Crop ratio left: ratio of pixels allow to be non blank while cutting on the left.
  -crop-ratio-up=int (default 1)
    	Crop ratio up: ratio of pixels allow to be non blank while cutting on the top.
  -crop-ratio-right=int (default 1)
    	Crop ratio right: ratio of pixels allow to be non blank while cutting on the right.
  -crop-ratio-bottom=int (default 3)
    	Crop ratio bottom: ratio of pixels allow to be non blank while cutting on the bottom.
  -crop-limit=int
    	Crop limit: maximum number of cropping in percentage allowed. 0 mean unlimited.
  -crop-skip-if-limit-reached or -crop-skip-if-limit-reached=1 (to enabled), -crop-skip-if-limit-reached=0 (to disabled)
    	Crop skip if limit reached.
  -brightness=int
    	Brightness readjustment: between -100 and 100, > 0 lighter, < 0 darker
  -contrast=int
    	Contrast readjustment: between -100 and 100, > 0 more contrast, < 0 less contrast
  -autocontrast or -autocontrast=1 (to enabled), -autocontrast=0 (to disabled)
    	Improve contrast automatically
  -doublepagemode=int
    	Double page mode (width > height)
    	0 = split
    	1 = split + rotate
    	2 = rotate
  -keepsplitdoublepageaspect or -keepsplitdoublepageaspect=1 (to enabled), -keepsplitdoublepageaspect=0 (to disabled) (default 1)
    	Keep aspect of split part of a double page (best for landscape rendering)
  -combinesplitdoublepage or -combinesplitdoublepage=1 (to enabled), -combinesplitdoublepage=0 (to disabled)
    	Detect a double page split in 2 consecutive pages, and combine them (the facing edges must match)
  -noblankimage or -noblankimage=1 (to enabled), -noblankimage=0 (to disabled) (default 1)
    	Remove blank image
  -manga or -manga=1 (to enabled), -manga=0 (to disabled)
    	Manga mode (right to left)
  -limitmb=int (default 200)
    	Limit size of the EPUB: Default limit (200), Minimum 20, 0 for unlimited
  -strip or -strip=1 (to enabled), -strip=0 (to disabled)
    	Strip first directory from the TOC if only 1
  -sort=int (default 1)
    	Sort path mode
    	0 = alpha for path and file
    	1 = alphanumeric for path and alpha for file
    	2 = alphanumeric for path and file
  -foreground-color=string (default "000000")
    	Foreground color in hexadecimal format RGB. Black=000000, White=FFFFFF
  -background-color=string (default "FFFFFF")
    	Background color in hexadecimal format RGB. Black=000000, White=FFFFFF, Light Gray=DDDDDD, Dark Gray=777777
  -resize or -resize=1 (to enabled), -resize=0 (to disabled) (default 1)
    	Reduce image size if exceed device size
  -format=string (default "jpeg")
    	Format of output images: jpeg (lossy), png (lossless), copy (no processing)
  -aspect-ratio=float
    	Aspect ratio (height/width) of the output, between 0.5 and 3
    	 -1 = same as device
    	  0 = same as source
    	1.6 = amazon advice for kindle
  -titlepage=int (default 1)
    	Title page
    	0 = never
    	1 = always
    	2 = only if epub is split
  -covercaption or -covercaption=1 (to enabled), -covercaption=0 (to disabled) (default 1)
    	Set cover caption when the epub is split
  -font=string
    	Path to a TrueType font (.ttf) for the cover caption and title page. Default: embedded Literata

Default config:
  -show
    	Show your default parameters
  -save
    	Save your parameters as default
  -reset
    	Reset your parameters to default

Shortcut:
  -nofilter
    	Deactivate all filters
  -maxquality
    	Max quality: color png + noresize
  -bestquality
    	Max quality: color jpg q100 + noresize
  -greatquality
    	Max quality: grayscale jpg q90 + noresize
  -goodquality
    	Max quality: grayscale jpg q90

Compatibility:
  -applebookcompatibility or -applebookcompatibility=1 (to enabled), -applebookcompatibility=0 (to disabled)
    	Apple book compatibility

Other:
  -workers=int (default 12)
    	Number of workers
  -dry
    	Dry run to show all options
  -quiet
    	Disable progress bar
  -json
    	Output progression and information in Json format
  -version
    	Show current and available version
  -help
    	Show this help message

```

# Credit

This project is largely inspired from KCC (Kindle Comic Converter). Thanks:
 - [ciromattia](https://github.com/ciromattia/kcc)
 - [darodi fork](https://github.com/darodi/kcc)

# UI

Thanks for UI contribution:
 - [manueldidonna / Comic2Books](https://github.com/manueldidonna/comic2books)
