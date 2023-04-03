# go-comic-converter

Convert CBZ/CBR/Dir into Epub for e-reader devices (Kindle Devices, ...)

My goal is to make a simple, crossplatform, and fast tool to convert comics into epub.

Epub is now support by Amazon through [SendToKindle](https://www.amazon.com/gp/sendtokindle/), by Email or by using the App. So I've made it simple to support the size limit constraint of those services.

# Installation

First ensure to have a working version of GO: [Installation](https://go.dev/doc/install)

Then install the last version of the tool:
```
$ go install github.com/celogeek/go-comic-converter/v2
```

To force install a specific version:
```
$ go install github.com/celogeek/go-comic-converter/v2@V2TAG
```

Example:
```
$ go install github.com/celogeek/go-comic-converter/v2@v2.0.1
```

Add GOPATH to your PATH
```
$ export PATH=$(go env GOPATH)/bin:$PATH
```

# Supported image files

The supported image files are jpeg and png from the sources.

The extensions can be: `jpg`, `jpeg`, `png`.

The case for extensions doesn't matter.

# Usage

## Convert directory

Convert every supported image files found in the input directory:

```
$ go-comic-converter -profile KS -input ~/Download/MyComic
```

By default it will output: ~/Download/MyComic.epub

## Convert CBZ, ZIP, CBR, RAR, PDF

Convert every supported image files found in the input directory:

```
$ go-comic-converter -profile KS -input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF]
```

By default it will output: ~/Download/MyComic.epub

## Convert with size limit

If you send your ePub through Amazon service, you have some size limitation:
  - Email  : 50Mb (including encoding, so 40Mb for RAW file)
  - App    : 50Mb
  - Website: 200Mb

You can split your file using the "-limitmb MB" option:

```
go-comic-converter -profile KS -input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF] -limitmb 200
```

If you have more than 1 file the output will be:
  - ~/Download/MyComic PART_01.epub
  - ~/Download/MyComic PART_02.epub
  - ...

The ePub include as a first page:
  - Title
  - Part NUM / TOTAL

If the total is above 1, then the title of the epub include:
  - Title [part/total]

## Dry run

If you want to preview what will be set during the convertion without running the conversion, then you can use the `-dry` option.

```
$ go-comic-converter -input ~/Downloads/mymanga.cbr -profile KS -auto -manga -limitmb 200 -dry
Go Comic Converter

Options:
    Input              : ~/Downloads/mymanga.cbr
    Output             : ~/Downloads/mymanga.epub
    Author             : GO Comic Converter
    Title              : mymanga
    Workers            : 8
    Profile            : KS - Kindle Scribe - 1860x2480 - 16 levels of gray
    Quality            : 85
    Crop               : true
    Brightness         : 0
    Contrast           : 0
    AutoRotate         : true
    AutoSplitDoublePage: true
    NoBlankPage        : false
    Manga              : true
    HasCover           : true
    AddPanelView       : false
    LimitMb            : 200 Mb
```

## Change default settings

### Show current default option
```
$ go-comic-converter -show

Go Comic Converter

Options:
    Profile            :
    Quality            : 85
    Crop               : true
    Brightness         : 0
    Contrast           : 0
    AutoRotate         : false
    AutoSplitDoublePage: false
    NoBlankPage        : false
    Manga              : false
    HasCover           : true
    AddPanelView       : false
    LimitMb            : nolimit
```

### Change default settings
```
$ go-comic-converter -manga -auto -profile KS -limitmb 200 -save

Go Comic Converter

Options:
    Profile            : KS - Kindle Scribe - 1860x2480 - 16 levels of gray
    Quality            : 85
    Crop               : true
    Brightness         : 0
    Contrast           : 0
    AutoRotate         : true
    AutoSplitDoublePage: true
    NoBlankPage        : false
    Manga              : true
    HasCover           : true
    AddPanelView       : false
    LimitMb            : 200 Mb

Saving to ~/.go-comic-converter.yaml
```

If you want to change a setting, you can change only one of them
```
$ go-comic-converter -manga=0 -save

Go Comic Converter

Options:
    Profile            : KS - Kindle Scribe - 1860x2480 - 16 levels of gray
    Quality            : 85
    Crop               : true
    Brightness         : 0
    Contrast           : 0
    AutoRotate         : true
    AutoSplitDoublePage: true
    NoBlankPage        : false
    Manga              : false
    HasCover           : true
    AddPanelView       : false
    LimitMb            : 200 Mb

Saving to ~/.go-comic-converter.yaml
```

### Check
You can test the command dry above like
```
$ go-comic-converter -input ~/Downloads/mymanga.cbr -dry
Go Comic Converter

Options:
    Input              : ~/Downloads/mymanga.cbr
    Output             : ~/Downloads/mymanga.epub
    Author             : GO Comic Converter
    Title              : mymanga
    Workers            : 8
    Profile            : KS - Kindle Scribe - 1860x2480 - 16 levels of gray
    Quality            : 85
    Crop               : true
    Brightness         : 0
    Contrast           : 0
    AutoRotate         : true
    AutoSplitDoublePage: true
    NoBlankPage        : false
    Manga              : true
    HasCover           : true
    AddPanelView       : false
    LimitMb            : 200 Mb
```

### Reset default
To reset all value to default:

```
$ go-comic-converter -reset
Go Comic Converter

Options:
    Profile            :
    Quality            : 85
    Crop               : true
    Brightness         : 0
    Contrast           : 0
    AutoRotate         : false
    AutoSplitDoublePage: false
    NoBlankPage        : false
    Manga              : false
    HasCover           : true
    AddPanelView       : false
    LimitMb            : nolimit

Reset default to ~/.go-comic-converter.yaml
```

# Help

```
$ go-comic-converter -h

Usage of go-comic-converter:

Output:
  -input string
    	Source of comic to convert: directory, cbz, zip, cbr, rar, pdf
  -output string
    	Output of the epub (directory or epub): (default [INPUT].epub)
  -author string (default "GO Comic Converter")
    	Author of the epub
  -title string
    	Title of the epub
  -workers int (default CPU)
    	Number of workers
  -dry
    	Dry run to show all options

Config:
  -profile string
    	Profile to use:
    	    - K1      (   600x670 ) -  4 levels of gray - Kindle 1
    	    - K11     ( 1072x1448 ) - 16 levels of gray - Kindle 11
    	    - K2      (   600x670 ) - 15 levels of gray - Kindle 2
    	    - K34     (   600x800 ) - 16 levels of gray - Kindle Keyboard/Touch
    	    - K578    (   600x800 ) - 16 levels of gray - Kindle
    	    - KDX     (  824x1000 ) - 16 levels of gray - Kindle DX/DXG
    	    - KPW     (  758x1024 ) - 16 levels of gray - Kindle Paperwhite 1/2
    	    - KV      ( 1072x1448 ) - 16 levels of gray - Kindle Paperwhite 3/4/Voyage/Oasis
    	    - KPW5    ( 1236x1648 ) - 16 levels of gray - Kindle Paperwhite 5/Signature Edition
    	    - KO      ( 1264x1680 ) - 16 levels of gray - Kindle Oasis 2/3
    	    - KS      ( 1860x2480 ) - 16 levels of gray - Kindle Scribe
    	    - KoMT    (   600x800 ) - 16 levels of gray - Kobo Mini/Touch
    	    - KoG     (  768x1024 ) - 16 levels of gray - Kobo Glo
    	    - KoGHD   ( 1072x1448 ) - 16 levels of gray - Kobo Glo HD
    	    - KoA     (  758x1024 ) - 16 levels of gray - Kobo Aura
    	    - KoAHD   ( 1080x1440 ) - 16 levels of gray - Kobo Aura HD
    	    - KoAH2O  ( 1080x1430 ) - 16 levels of gray - Kobo Aura H2O
    	    - KoAO    ( 1404x1872 ) - 16 levels of gray - Kobo Aura ONE
    	    - KoN     (  758x1024 ) - 16 levels of gray - Kobo Nia
    	    - KoC     ( 1072x1448 ) - 16 levels of gray - Kobo Clara HD/Kobo Clara 2E
    	    - KoL     ( 1264x1680 ) - 16 levels of gray - Kobo Libra H2O/Kobo Libra 2
    	    - KoF     ( 1440x1920 ) - 16 levels of gray - Kobo Forma
    	    - KoS     ( 1440x1920 ) - 16 levels of gray - Kobo Sage
    	    - KoE     ( 1404x1872 ) - 16 levels of gray - Kobo Elipsa
  -quality int (default 85)
    	Quality of the image
  -crop (default true)
    	Crop images
  -brightness int
    	Brightness readjustement: between -100 and 100, > 0 lighter, < 0 darker
  -contrast int
    	Contrast readjustement: between -100 and 100, > 0 more contrast, < 0 less contrast
  -autorotate
    	Auto Rotate page when width > height
  -auto
    	Activate all automatic options
  -autosplitdoublepage
    	Auto Split double page when width > height
  -noblankpage
    	Remove blank pages
  -manga
    	Manga mode (right to left)
  -hascover (default true)
    	Has cover. Indicate if your comic have a cover. The first page will be used as a cover and include after the title.
  -addpanelview
    	Add an embeded panel view. On kindle you may not need this option as it is handled by the kindle.
  -limitmb int
    	Limit size of the ePub: Default nolimit (0), Minimum 20

Default config:
  -show
    	Show your default parameters
  -save
    	Save your parameters as default
  -reset
    	Reset your parameters to default

Other:
  -help
    	Show this help message
```

# Credit

This project is largely inspired from KCC (Kindle Comic Converter). Thanks:
 - [ciromattia](https://github.com/ciromattia/kcc)
 - [darodi fork](https://github.com/darodi/kcc)

