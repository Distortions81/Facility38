package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const fpx = 96.0

func updateFonts() {
	defer reportPanic("updateFonts")

	newVal := fpx * UIScale
	if newVal < 1 {
		newVal = 1
	}
	FontDPI = newVal

	now := time.Now()
	var mono, tt *opentype.Font
	var logo *opentype.Font
	var err error

	if now.Month() == 4 && now.Day() == 1 {
		fdata := GetFont("comici.ttf")
		collection, err := opentype.ParseCollection(fdata)
		if err != nil {
			log.Fatal(err)
		}

		tt, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fdata := GetFont("Exo2-Regular.ttf")
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
	fdata := GetFont("Azonix-1VB0.otf")
	collection, err := opentype.ParseCollection(fdata)
	if err != nil {
		log.Fatal(err)
	}

	logo, err = collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	/* Mono font */
	fdata = GetFont("Hack-Regular.ttf")
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
	BootFont, err = opentype.NewFace(logo, &opentype.FaceOptions{
		Size:    25,
		DPI:     FontDPI,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	BootFontH = getFontHeight(BootFont)

	/* General font */
	GeneralFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    10,
		DPI:     FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	GeneralFontH = getFontHeight(GeneralFont)

	/* Missing texture font */
	ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    6,
		DPI:     FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	ObjectFontH = getFontHeight(ObjectFont)

	/* Tooltip font */
	ToolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    8,
		DPI:     FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	ToolTipFontH = getFontHeight(ToolTipFont)

	/* Mono font */
	MonoFont, err = opentype.NewFace(mono, &opentype.FaceOptions{
		Size:    8,
		DPI:     FontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	MonoFontH = getFontHeight(MonoFont)

	/* Logo font */
	LogoFont, err = opentype.NewFace(logo, &opentype.FaceOptions{
		Size:    70,
		DPI:     FontDPI,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	LogoFontH = getFontHeight(LogoFont)
}

const sizingText = "!@#$%^&*()_+-=[]{}|;':,.<>?`~qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

func getFontHeight(font font.Face) int {
	defer reportPanic("getFontHeight")
	tRect := text.BoundString(font, sizingText)
	return tRect.Dy()
}
