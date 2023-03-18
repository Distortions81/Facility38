package main

import (
	"GameTest/data"
	"GameTest/world"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

/* Font setup, eventually use ttf files */
func init() {

	fdata := data.GetFont("Manjari-Bold.otf")
	collection, err := opentype.ParseCollection(fdata)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	fdata = data.GetFont("NotoSansCJK-Bold.ttc")
	collection, err = opentype.ParseCollection(fdata)
	if err != nil {
		log.Fatal(err)
	}

	mono, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * Font DPI
	 * Changes how large the font is for a given point value
	 */
	const dpi = 96
	/* Boot screen font */
	world.BootFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    15,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Missing texture font */
	world.ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    5,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Tooltip font */
	world.ToolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    11,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Mono font */
	world.MonoFont, err = opentype.NewFace(mono, &opentype.FaceOptions{
		Size:    11,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

}
