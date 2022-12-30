# go-comic-converter

Convert CBZ/CBR/Dir into Epub for e-reader devices (Kindle Devices, ...)

# Installation

First ensure to have a working version of GO: [Installation](https://go.dev/doc/install)

Then install the last version of the tool:
```
go install github.com/celogeek/go-comic-converter@latest
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
    	    - K1      (   600x670 ) - Kindle 1
    	    - K11     ( 1072x1448 ) - Kindle 11
    	    - K2      (   600x670 ) - Kindle 2
    	    - K34     (   600x800 ) - Kindle Keyboard/Touch
    	    - K578    (   600x800 ) - Kindle
    	    - KDX     (  824x1000 ) - Kindle DX/DXG
    	    - KPW     (  758x1024 ) - Kindle Paperwhite 1/2
    	    - KV      ( 1072x1448 ) - Kindle Paperwhite 3/4/Voyage/Oasis
    	    - KPW5    ( 1236x1648 ) - Kindle Paperwhite 5/Signature Edition
    	    - KO      ( 1264x1680 ) - Kindle Oasis 2/3
    	    - KS      ( 1860x2480 ) - Kindle Scribe
    	    - KoMT    (   600x800 ) - Kobo Mini/Touch
    	    - KoG     (  768x1024 ) - Kobo Glo
    	    - KoGHD   ( 1072x1448 ) - Kobo Glo HD
    	    - KoA     (  758x1024 ) - Kobo Aura
    	    - KoAHD   ( 1080x1440 ) - Kobo Aura HD
    	    - KoAH2O  ( 1080x1430 ) - Kobo Aura H2O
    	    - KoAO    ( 1404x1872 ) - Kobo Aura ONE
    	    - KoN     (  758x1024 ) - Kobo Nia
    	    - KoC     ( 1072x1448 ) - Kobo Clara HD/Kobo Clara 2E
    	    - KoL     ( 1264x1680 ) - Kobo Libra H2O/Kobo Libra 2
    	    - KoF     ( 1440x1920 ) - Kobo Forma
    	    - KoS     ( 1440x1920 ) - Kobo Sage
    	    - KoE     ( 1404x1872 ) - Kobo Elipsa
  -quality int
    	Quality of the image (default 85)
  -title string
    	Title of the epub
```