package main

import (
	"Facility38/data"
	"Facility38/util"
	"Facility38/world"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func UpdateFonts() {
	defer util.ReportPanic("UpdateFonts")

	newVal := 96.0 * world.UIScale
	if newVal < 1 {
		newVal = 1
	}
	world.FontDPI = newVal

	now := time.Now()
	var mono, tt *opentype.Font
	var logo *opentype.Font
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
	} else {
		fdata := data.GetFont("Exo2-Regular.ttf")
		collection, err := opentype.ParseCollection(fdata)
		if err != nil {
			log.Fatal(err)
		}

		tt, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}
	}

	/* Logo font */
	fdata := data.GetFont("Azonix-1VB0.otf")
	collection, err := opentype.ParseCollection(fdata)
	if err != nil {
		log.Fatal(err)
	}

	logo, err = collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	fdata = data.GetFont("Hack-Regular.ttf")
	collection, err = opentype.ParseCollection(fdata)
	if err != nil {
		log.Fatal(err)
	}

	mono, err = collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * Font DPI
	 * Changes how large the font is for a given point value
	 */

	/* Boot screen font */
	world.BootFont, err = opentype.NewFace(logo, &opentype.FaceOptions{
		Size:    25,
		DPI:     world.FontDPI,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	world.BootFontH = getFontHeight(world.BootFont)

	/* General font */
	world.GeneralFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    10,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	world.GeneralFontH = getFontHeight(world.GeneralFont)

	/* Missing texture font */
	world.ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    6,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	world.ObjectFontH = getFontHeight(world.ObjectFont)

	/* Tooltip font */
	world.ToolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    8,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	world.ToolTipFontH = getFontHeight(world.ToolTipFont)

	/* Mono font */
	world.MonoFont, err = opentype.NewFace(mono, &opentype.FaceOptions{
		Size:    8,
		DPI:     world.FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	world.MonoFontH = getFontHeight(world.MonoFont)

	/* Logo font */
	world.LogoFont, err = opentype.NewFace(logo, &opentype.FaceOptions{
		Size:    70,
		DPI:     world.FontDPI,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	world.LogoFontH = getFontHeight(world.LogoFont)
}

const sizingText = "!@#$%^&*()_+-=[]{}|;':,.<>?`~qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

func getFontHeight(font font.Face) int {
	defer util.ReportPanic("getFontHeight")
	tRect := text.BoundString(font, sizingText)
	return tRect.Dy()
}
