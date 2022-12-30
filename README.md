# go-comic-converter

Convert CBZ/CBR/Dir into Epub for e-reader devices (Kindle Devices, ...)

My goal is to make a simple, crossplatform, and fast tool to convert comics into epub.

Epub is now support by Amazon through [SendToKindle](https://www.amazon.com/gp/sendtokindle/), by Email or by using the App. So I've made it simple to support the size limit constraint of those services.

# Installation

First ensure to have a working version of GO: [Installation](https://go.dev/doc/install)

Then install the last version of the tool:
```
go install github.com/celogeek/go-comic-converter@latest
```

To force install a specific version:
```
go install github.com/celogeek/go-comic-converter@TAG
# Ex: go install github.com/celogeek/go-comic-converter@v1.0.0
```

Add GOPATH to your PATH
```
export PATH=$(go env GOPATH)/bin:$PATH
```

# Usage

## Convert directory

Convert every ".jpg" file found in the input directory:

```
go-comic-converter --profile KS --input ~/Download/MyComic
```

By default it will output: ~/Download/MyComic.epub

## Convert CBZ, ZIP, CBR, RAR, PDF

Convert every ".jpg" file found in the input directory:

```
go-comic-converter --profile KS --input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF]
```

By default it will output: ~/Download/MyComic.epub

## Convert with size limit

If you send your ePub through Amazon service, you have some size limitation:
  - Email  : 50Mb (including encoding, so 40Mb for RAW file)
  - App    : 50Mb
  - Website: 200Mb

You can split your file using the "--limitmb MB" option:

```
go-comic-converter --profile KS --input ~/Download/MyComic.[CBZ,ZIP,CBR,RAR,PDF] --limitmb 200
```

If you have more than 1 file the output will be:
  - ~/Download/MyComic PART_01.epub
  - ~/Download/MyComic PART_02.epub
  - ...

The ePub include as a first page:
  - Title
  - Part NUM / TOTAL

# Help

```
# go-comic-converter -h

Usage of go-comic-converter:
  -algo string
    	Algo for RGB to Grayscale: default, mean, luma, luster (default "default")
  -author string
    	Author of the epub (default "GO Comic Converter")
  -input string
    	Source of comic to convert: directory, cbz, zip, cbr, rar, pdf
  -limitmb int
    	Limit size of the ePub: Default nolimit (0), Minimum 20
  -nocrop
    	Disable cropping
  -output string
    	Output of the epub (directory or epub): (default [INPUT].epub)
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
  -quality int
    	Quality of the image (default 85)
  -title string
    	Title of the epub
```

# Credit

This project is largely inspired from KCC (Kindle Comic Converter). Thanks:
 - [ciromattia](https://github.com/ciromattia/kcc)
 - [darodi fork](https://github.com/darodi/kcc)

