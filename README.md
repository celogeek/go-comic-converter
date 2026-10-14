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
- Images are tiled only for the devices with a limit on the size of an image: Kindle (1200x1920, the SR resolution) and Apple Books (2288x2432, under the 5.6 million pixels limit).
  A bigger image is stored as a grid of tiles shown as one page, to keep the resolution on large devices. Other devices (Kobo, reMarkable) are never tiled
- Zoom of the stored images (`-zoom`): they are stored bigger than the device (1.5 for Apple Books, as advised by Apple) for a better quality when you zoom
- Optionally detect a double page split in 2 consecutive pages, and combine them (`-combinesplitdoublepage`, off by default)
- Remove blank image (empty image is removed)
- Manga or Normal mode
- Landscape only mode for Kindle Scribe (`-landscape`, off by default): the pages are spread left and right, with a blank page to keep the double pages facing each other
- Embedded panel view for the older devices that need it (set by the profile)
- The cover is the first image that is not blank
- An unreadable image is replaced by a page with its name, instead of stopping the conversion
- Support title page (cover with embedded title and part)
- Split EPUB size for easy upload
- 3 sorting methods (depending on your source, you can ensure the page go in the right order)
- Save and reuse your own perfect settings
- Multi tasks for fast conversion
- Apple Books mode, set by the profile: the double page is split and keeps its aspect, the spread follows Apple Books, and the tile size follows its limit (`AB` profile)
- JSON output for programmatic usage
- Support cover caption with the part number when the EPUB is split
- Support for own font for the title page and the cover caption

When you read the comic on a Kindle, you can customize how you read it with the `Aa` button:
- Landscape / Portrait
- Activate panel view for small device

The panel view is handled natively by recent Kindles, for example the Kindle Scribe.
The embedded panel view comes only from the profile: it is enabled for the older devices that need it, and can't be changed from the command line or the saved config.

The grayscale can only be turned off (`-grayscale=0`) with a profile of a color device (the `Color` column of the profiles, see `-profile`).

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
    Profile                         : SR - Kindle Standard Resolution - 1200x1920
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
    Profile                         : SR - Kindle Standard Resolution - 1200x1920
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
    Profile                         : SR - Kindle Standard Resolution - 1200x1920
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
    Profile                         : SR - Kindle Standard Resolution - 1200x1920
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
    Profile                         : SR - Kindle Standard Resolution - 1200x1920
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

### Ignore the saved default
The help and the options start from the parameters you saved. To ignore them for one run, and use the built-in default:

```
$ go-comic-converter -help -ignore-config
$ go-comic-converter -input MyComic.cbz -ignore-config
```

Your saved file is not changed, unless you add `-save` (the built-in default, and the other options you set, are saved).

### Show the command line
To get a command line that gives the same result whatever the saved parameters of the one who runs it
(to share it, or to keep it), use `-show-command`:

```
$ go-comic-converter -profile AB -manga -quality 70 -input MyComic.cbz -show-command
go-comic-converter -ignore-config \
  -input=MyComic.cbz \
  -profile=AB \
  -quality=70 \
  -manga=1
```

It starts with `-ignore-config`, so the saved parameters of the one who runs it don't matter, then the profile,
and every option that differs from the built-in default. The options at their default are skipped, and so is the zoom of the profile
(1.5 for Apple Books, 1 for the others). The options you saved are part of it, add `-ignore-config` to your own arguments to leave them out.

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
    	Device      Code     Year  Resolution   Zoom  Color  Panel view  Landscape  Description
    	----------  -------  ----  -----------  ----  -----  ----------  ---------  --------------------------------------
    	Kindle      HR             2400 x 3840  x1    yes                           Kindle High Resolution
    	Kindle      SR             1200 x 1920  x1    yes                           Kindle Standard Resolution
    	Kindle      KS3      2025  1980 x 2640  x1                       yes        Kindle Scribe 3
    	Kindle      KSC      2025  1980 x 2640  x1    yes                yes        Kindle Scribe Colorsoft
    	Kindle      K12      2024  1072 x 1448  x1                                  Kindle 12
    	Kindle      KPW12    2024  1264 x 1680  x1                                  Kindle Paperwhite 12/Signature Edition
    	Kindle      KCS      2024  1264 x 1680  x1    yes                           Kindle Colorsoft
    	Kindle      KS2      2024  1860 x 2480  x1                       yes        Kindle Scribe 2
    	Kindle      K11      2022  1072 x 1448  x1                                  Kindle 11
    	Kindle      KS       2022  1860 x 2480  x1                       yes        Kindle Scribe
    	Kindle      KPW5     2021  1236 x 1648  x1           yes                    Kindle Paperwhite 5/Signature Edition
    	Kindle      K10      2019   600 x 800   x1           yes                    Kindle 10
    	Kindle      KO3      2019  1264 x 1680  x1           yes                    Kindle Oasis 3
    	Kindle      KPW4     2018  1072 x 1448  x1           yes                    Kindle Paperwhite 4
    	Kindle      KO2      2017  1264 x 1680  x1           yes                    Kindle Oasis 2
    	Kindle      KO1      2016  1072 x 1448  x1           yes                    Kindle Oasis 1
    	Kindle      K8       2016   600 x 800   x1           yes                    Kindle 8
    	Kindle      KPW3     2015  1072 x 1448  x1           yes                    Kindle Paperwhite 3
    	Kindle      K7       2014   600 x 800   x1           yes                    Kindle 7
    	Kindle      KV       2014  1072 x 1448  x1           yes                    Kindle Voyage
    	Kindle      KPW2     2013   758 x 1024  x1           yes                    Kindle Paperwhite 2
    	Kindle      K5       2012   600 x 800   x1           yes                    Kindle 5
    	Kindle      KPW1     2012   758 x 1024  x1           yes                    Kindle Paperwhite 1
    	Kindle      K4       2011   600 x 800   x1           yes                    Kindle 4
    	Kindle      KT       2011   600 x 800   x1           yes                    Kindle Touch
    	Kindle      K3       2010   600 x 800   x1           yes                    Kindle Keyboard
    	Kindle      K2       2009   600 x 800   x1           yes                    Kindle 2
    	Kindle      KDX      2009   824 x 1200  x1           yes                    Kindle DX/DXG
    	Kindle      K1       2007   600 x 800   x1           yes                    Kindle 1
    	Kobo        KoCBW    2024  1072 x 1448  x1                                  Kobo Clara BW
    	Kobo        KoCC     2024  1072 x 1448  x1    yes                           Kobo Clara Colour
    	Kobo        KoLC     2024  1264 x 1680  x1    yes                           Kobo Libra Colour
    	Kobo        KoE2     2023  1404 x 1872  x1                                  Kobo Elipsa 2E
    	Kobo        KoC2E    2022  1072 x 1448  x1                                  Kobo Clara 2E
    	Kobo        KoE      2021  1404 x 1872  x1                                  Kobo Elipsa
    	Kobo        KoS      2021  1440 x 1920  x1                                  Kobo Sage
    	Kobo        KoL2     2021  1264 x 1680  x1                                  Kobo Libra 2
    	Kobo        KoN      2020   758 x 1024  x1                                  Kobo Nia
    	Kobo        KoLH2O   2019  1264 x 1680  x1                                  Kobo Libra H2O
    	Kobo        KoCHD    2018  1072 x 1448  x1                                  Kobo Clara HD
    	Kobo        KoF      2018  1440 x 1920  x1                                  Kobo Forma
    	Kobo        KoAH2O2  2017  1080 x 1430  x1                                  Kobo Aura H2O Edition 2
    	Kobo        KoAO     2016  1404 x 1872  x1                                  Kobo Aura One
    	Kobo        KoA2     2016   758 x 1024  x1                                  Kobo Aura Edition 2
    	Kobo        KoGHD    2015  1072 x 1448  x1                                  Kobo Glo HD
    	Kobo        KoT2     2015   600 x 800   x1                                  Kobo Touch 2.0
    	Kobo        KoAH2O   2014  1080 x 1430  x1                                  Kobo Aura H2O
    	Kobo        KoA      2013   758 x 1024  x1                                  Kobo Aura
    	Kobo        KoAHD    2013  1080 x 1440  x1                                  Kobo Aura HD
    	Kobo        KoM      2012   600 x 800   x1                                  Kobo Mini
    	Kobo        KoG      2012   758 x 1024  x1                                  Kobo Glo
    	Kobo        KoT      2011   600 x 800   x1                                  Kobo Touch
    	Kobo        KoR      2010   600 x 800   x1                                  Kobo eReader
    	reMarkable  RMPU     2026  1404 x 1872  x1                                  reMarkable Paper Pure
    	reMarkable  RMPM     2025   954 x 1696  x1    yes                           reMarkable Paper Pro Move
    	reMarkable  RMP      2024  1620 x 2160  x1    yes                           reMarkable Paper Pro
    	reMarkable  RM2      2020  1404 x 1872  x1                                  reMarkable 2
    	reMarkable  RM1      2017  1404 x 1872  x1                                  reMarkable 1
    	Apple       AB             1640 x 2360  x1.5  yes                           Apple Books Standard Resolution
    	Apple       ABD      2026  1920 x 2713  x1.5  yes                           Apple Books on iPhone Duo
    	Apple       ABP11    2025  1668 x 2420  x1.5  yes                           Apple Books on iPad Pro 11"
    	Apple       ABP13    2025  2064 x 2752  x1.5  yes                           Apple Books on iPad Pro 13"
    	Apple       ABI      2025  1320 x 2868  x1.5  yes                           Apple Books on iPhone
    	Apple       ABM      2024  1488 x 2266  x1.5  yes                           Apple Books on iPad mini
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
  -landscape or -landscape=1 (to enabled), -landscape=0 (to disabled)
    	Landscape only mode, Kindle Scribe only (an error with the other profiles): the pages are spread left and right, with a blank page to keep the double pages facing each other
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
  -zoom=float
    	Zoom of the images, between 1 and 3: they are stored bigger than the device for a better quality when you zoom
    	  0 = zoom of the profile (1.5 for Apple Books, 1 for the others)
    	  1 = size of the device
    	The size of the EPUB grows with the square of the zoom
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
  -show-command
    	Show the command line equivalent to the current options: -ignore-config, the profile, and the options that differ from the default
  -ignore-config
    	Ignore your saved parameters, and use the default ones (with -help to show them)

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
