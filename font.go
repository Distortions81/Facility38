package main

import (
	"Facility38/data"
	"Facility38/gv"
	"Facility38/world"
	"log"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func UpdateFonts() {

	newVal := 96.0 * gv.UIScale
	if newVal < 1 {
		newVal = 1
	} else if newVal > 600 {
		newVal = 600
	}
	world.FontDPI = newVal

	now := time.Now()
	var mono, tt *opentype.Font
	var err error

	if now.Month() == 4 && now.Day() == 1 {
		fdata := data.GetFont("comici.ttf")
		collection, err := opentype.ParseCollection(fdata)
		if err != nil {
			log.Fatal(err)
		}

		tt, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}

		mono = tt
	} else {

		fdata := data.GetFont("Manjari-Bold.otf")
		collection, err := opentype.ParseCollection(fdata)
		if err != nil {
			log.Fatal(err)
		}

		tt, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}

		fdata = data.GetFont("NotoSansMono-Bold.ttf")
		collection, err = opentype.ParseCollection(fdata)
		if err != nil {
			log.Fatal(err)
		}

		mono, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}
	}

	/*
	 * Font DPI
	 * Changes how large the font is for a given point value
	 */
	/* Boot screen font */
	world.BootFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    15,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Missing texture font */
	world.ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    5,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Tooltip font */
	world.ToolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    11,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Mono font */
	world.MonoFont, err = opentype.NewFace(mono, &opentype.FaceOptions{
		Size:    11,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}
