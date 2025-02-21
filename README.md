# go-comic-converter

Convert CBZ/CBR/Dir into EPUB for e-reader devices (Kindle Devices, ...)

My goal is to make a simple, cross-platform, and fast tool to convert comics into EPUB.

EPUB is now support by Amazon through [SendToKindle](https://www.amazon.com/gp/sendtokindle/), by Email or by using the App. So I've made it simple to support the size limit constraint of those services.

# Features
- Support input from zip, cbz, rar, cbr, pdf, directory
- Support all Kindle devices and kobo
- Support Landscape and Portrait mode
- Customize output image quality
- Intelligent cropping (support removing even page numbers)
- Customize brightness and contrast
- Auto contrast
- Auto rotate (if reader mainly read on portrait)
- Auto split double page (for easy read on portrait)
- Keep double page if split
- Remove blank image (empty image is removed)
- Manga or Normal mode
- Support cover page or not (first page will be taken in that case)
- Support title page (cover with embedded title and part)
- Split EPUB size for easy upload
- 3 sorting methods (depending on your source, you can ensure the page go in the right order)
- Save and reuse your own perfect settings
- Multi tasks for fast conversion
- Apple Book Compatibility Mode
- JSON output for programmatic usage

When you read the comic on a Kindle, you can customize how you read it with the `Aa` button:
- Landscape / Portrait
- Activate panel view for small device

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

You can split your file using the "-limitmb MB" option:

```
go-comic-converter -profile SR -input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF] -limitmb 200
```

If you have more than 1 file the output will be:
  - ~/Download/MyComic Part 01 of 03.epub
  - ~/Download/MyComic Part 02 of 03.epub
  - ...

The ePub include as a first page:
  - Title
  - Part NUM / TOTAL

If the total is above 1, then the title of the EPUB include:
  - Title [part/total]

## Dry run

If you want to preview what will be set during the conversion without running the conversion, then you can use the `-dry` option.

```
$ go-comic-converter -input ~/Downloads/mymanga.cbr -profile SR -auto -manga -limitmb 200 -dry
Go Comic Converter

Options:
    Input                           : ~/Downloads/mymanga.cbr
    Output                          : ~/Downloads/mymanga.epub
    Author                          : GO Comic Converter
    Title                           : mymanga
    Workers                         : 20
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : true
    Auto rotate                     : true
    Auto split double page          : true
    Keep double page if split       : true
    No blank image                  : true
    Manga                           : true
    Has cover                       : true
    Limit                           : 200 Mb
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : auto
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false

TOC:
  - mymanga
  - Chapter 1
  - Chapter 2
  - Chapter 3
```

## Dry verbose

You can choose different way to sort path and files, depending on your source. You can preview the sorted result with the option `dry-verbose` associated with `dry`.

The option `sort` allow you to change the sorting order.

```
$ go-comic-converter -input ~/Downloads/mymanga.cbr -profile SR -auto -manga -limitmb 200 -dry -dry-verbose -sort 2
Go Comic Converter

Options:
    Input                           : ~/Downloads/mymanga.cbr
    Output                          : ~/Downloads/mymanga.epub
    Author                          : GO Comic Converter
    Title                           : mymanga
    Workers                         : 20
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : true
    Auto rotate                     : true
    Auto split double page          : true
    Keep double page if split       : true
    No blank image                  : true
    Manga                           : true
    Has cover                       : true
    Limit                           : 200 Mb
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : auto
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false

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
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : false
    Auto rotate                     : false
    Auto split double page          : false
    No blank image                  : true
    Manga                           : false
    Has cover                       : true
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : auto
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false
```

### Change default settings
```
$ go-comic-converter -manga -auto -profile SR -limitmb 200 -save

Go Comic Converter

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : true
    Auto rotate                     : true
    Auto split double page          : true
    Keep double page if split       : true
    Keep split double page aspect   : true
    No blank image                  : true
    Manga                           : true
    Has cover                       : true
    Limit                           : 200 Mb
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : auto
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false

Saving to ~/.go-comic-converter.yaml
```

If you want to change a setting, you can change only one of them
```
$ go-comic-converter -manga=0 -save

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 85
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : false
    Auto rotate                     : false
    Auto split double page          : false
    No blank image                  : true
    Manga                           : false
    Has cover                       : true
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : auto
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false

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
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : false
    Auto rotate                     : false
    Auto split double page          : false
    No blank image                  : true
    Manga                           : false
    Has cover                       : true
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : auto
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false

Reset default to ~/.go-comic-converter.yaml
```

# My own settings

After playing around with the options, I have my perfect settings for all my devices.

```
$ go-comic-converter -reset
$ go-comic-converter -profile SR -quality 90 -manga -aspect-ratio 1.6 -limitmb 200 -save

Options:
    Profile                         : SR - Standard Resolution - 1200x1920
    Format                          : jpeg
    Quality                         : 90
    Grayscale                       : true
    Grayscale mode                  : normal
    Crop                            : true
    Crop ratio                      : 1 Left - 1 Up - 1 Right - 3 Bottom - Limit 0% - Skip false
    Auto contrast                   : false
    Auto rotate                     : false
    Auto split double page          : false
    No blank image                  : true
    Manga                           : true
    Has cover                       : true
    Limit                           : 200 Mb
    Strip first directory from toc  : false
    Sort path mode                  : path=alphanumeric, file=alpha
    Foreground color                : #000
    Background color                : #FFF
    Resize                          : true
    Aspect ratio                    : 1:1.60
    Portrait only                   : false
    Title page                      : always
    Apple book compatibility        : false

Saving to ~/.go-comic-converter.yaml
```

Explanation:
- `-profile SR`: standard resolution (fast conversion from Amazon as images do not need to be resized)
- `-quality 90`: JPEG output quality of images
- `-manga`: manga mode, read right to left
- `-limitmb 200`: size limit to 200MB allowing upload from SendToKindle website
- `-aspect-ratio`: ensure aspect ratio is 1:1.6, best for kindle devices.
# Help

```
$ go-comic-converter -h

Usage of go-comic-converter:

Output:
  -input string
    	Source of comic to convert: directory, cbz, zip, cbr, rar, pdf
  -output string
    	Output of the EPUB (directory or EPUB): (default [INPUT].epub)
  -author string (default "GO Comic Converter")
    	Author of the EPUB
  -title string
    	Title of the EPUB

Config:
  -profile string (default "SR")
    	Profile to use: 
    	    - KoAO    - 1404 x 1872 - Kobo Aura ONE
    	    - KoF     - 1440 x 1920 - Kobo Forma
    	    - KoE     - 1404 x 1872 - Kobo Elipsa
    	    - KV      - 1072 x 1448 - Kindle Paperwhite 3/4/Voyage/Oasis
    	    - KoG     -  768 x 1024 - Kobo Glo
    	    - KoA     -  758 x 1024 - Kobo Aura
    	    - RM1     - 1404 x 1872 - reMarkable 1
    	    - RM2     - 1404 x 1872 - reMarkable 2
    	    - K1      -  600 x 670  - Kindle 1
    	    - K11     - 1072 x 1448 - Kindle 11
    	    - K2      -  600 x 670  - Kindle 2
    	    - K34     -  600 x 800  - Kindle Keyboard/Touch
    	    - KPW5    - 1236 x 1648 - Kindle Paperwhite 5/Signature Edition
    	    - KoAH2O  - 1080 x 1430 - Kobo Aura H2O
    	    - KoN     -  758 x 1024 - Kobo Nia
    	    - KoL     - 1264 x 1680 - Kobo Libra H2O/Kobo Libra 2
    	    - HR      - 2400 x 3840 - High Resolution
    	    - KO      - 1264 x 1680 - Kindle Oasis 2/3
    	    - KS      - 1860 x 2480 - Kindle Scribe
    	    - KoMT    -  600 x 800  - Kobo Mini/Touch
    	    - KoAHD   - 1080 x 1440 - Kobo Aura HD
    	    - KoC     - 1072 x 1448 - Kobo Clara HD/Kobo Clara 2E
    	    - KoS     - 1440 x 1920 - Kobo Sage
    	    - SR      - 1200 x 1920 - Standard Resolution
    	    - K578    -  600 x 800  - Kindle
    	    - KDX     -  824 x 1000 - Kindle DX/DXG
    	    - KPW     -  758 x 1024 - Kindle Paperwhite 1/2
    	    - KoGHD   - 1072 x 1448 - Kobo Glo HD
  -quality int (default 85)
    	Quality of the image
  -grayscale (default true)
    	Grayscale image. Ideal for eInk devices.
  -grayscale-mode int
    	Grayscale Mode
    	0 = normal
    	1 = average
    	2 = luminance
  -crop (default true)
    	Crop images
  -crop-ratio-left int (default 1)
    	Crop ratio left: ratio of pixels allow to be non blank while cutting on the left.
  -crop-ratio-up int (default 1)
    	Crop ratio up: ratio of pixels allow to be non blank while cutting on the top.
  -crop-ratio-right int (default 1)
    	Crop ratio right: ratio of pixels allow to be non blank while cutting on the right.
  -crop-ratio-bottom int (default 3)
    	Crop ratio bottom: ratio of pixels allow to be non blank while cutting on the bottom.
  -crop-limit int
    	Crop limit: maximum number of cropping in percentage allowed. 0 mean unlimited.
  -crop-skip-if-limit-reached
    	Crop skip if limit reached.
  -brightness int
    	Brightness readjustment: between -100 and 100, > 0 lighter, < 0 darker
  -contrast int
    	Contrast readjustment: between -100 and 100, > 0 more contrast, < 0 less contrast
  -autocontrast
    	Improve contrast automatically
  -autorotate
    	Auto Rotate page when width > height
  -autosplitdoublepage
    	Auto Split double page when width > height
  -keepdoublepageifsplit (default true)
    	Keep the double page if split
  -keepsplitdoublepageaspect (default true)
    	Keep aspect of split part of a double page (best for landscape rendering)
  -noblankimage (default true)
    	Remove blank image
  -manga
    	Manga mode (right to left)
  -hascover (default true)
    	Has cover. Indicate if your comic have a cover. The first page will be used as a cover and include after the title.
  -limitmb int
    	Limit size of the EPUB: Default nolimit (0), Minimum 20
  -strip
    	Strip first directory from the TOC if only 1
  -sort int (default 1)
    	Sort path mode
    	0 = alpha for path and file
    	1 = alphanumeric for path and alpha for file
    	2 = alphanumeric for path and file
  -foreground-color string (default "000")
    	Foreground color in hexadecimal format RGB. Black=000, White=FFF
  -background-color string (default "FFF")
    	Background color in hexadecimal format RGB. Black=000, White=FFF, Light Gray=DDD, Dark Gray=777
  -resize (default true)
    	Reduce image size if exceed device size
  -format string (default "jpeg")
    	Format of output images: jpeg (lossy), png (lossless), copy (no processing)
  -aspect-ratio float
    	Aspect ratio (height/width) of the output
    	 -1 = same as device
    	  0 = same as source
    	1.6 = amazon advice for kindle
  -portrait-only
    	Portrait only: force orientation to portrait only.
  -titlepage int (default 1)
    	Title page
    	0 = never
    	1 = always
    	2 = only if epub is split

Default config:
  -show
    	Show your default parameters
  -save
    	Save your parameters as default
  -reset
    	Reset your parameters to default

Shortcut:
  -auto
    	Activate all automatic options
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
  -applebookcompatibility
    	Apple book compatibility

Other:
  -workers int (default number of CPUs)
    	Number of workers
  -dry
    	Dry run to show all options
  -dry-verbose
    	Display also sorted files after the TOC
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
