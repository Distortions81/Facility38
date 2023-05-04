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

	newVal := fpx * uiScale
	if newVal < 1 {
		newVal = 1
	}
	fontDPI = newVal

	now := time.Now()
	var mono, tt *opentype.Font
	var logo *opentype.Font
	var err error

	if now.Month() == 4 && now.Day() == 1 {
		fontData := getFont("comici.ttf")
		collection, err := opentype.ParseCollection(fontData)
		if err != nil {
			log.Fatal(err)
		}

		tt, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fontData := getFont("Exo2-Regular.ttf")
		collection, err := opentype.ParseCollection(fontData)
		if err != nil {
			log.Fatal(err)
		}

		tt, err = collection.Font(0)
		if err != nil {
			log.Fatal(err)
		}
	}

	/* Logo font */
	fontData := getFont("Azonix-1VB0.otf")
	collection, err := opentype.ParseCollection(fontData)
	if err != nil {
		log.Fatal(err)
	}

	logo, err = collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	/* Mono font */
	fontData = getFont("Hack-Regular.ttf")
	collection, err = opentype.ParseCollection(fontData)
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
	bootFont, err = opentype.NewFace(logo, &opentype.FaceOptions{
		Size:    25,
		DPI:     fontDPI,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	bootFontH = getFontHeight(bootFont)

	/* General font */
	generalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    10,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	generalFontH = getFontHeight(generalFont)

	/* Missing texture font */
	objectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    6,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	objectFontH = getFontHeight(objectFont)

	/* Tooltip font */
	toolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    8,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	toolTipFontH = getFontHeight(toolTipFont)

	/* Mono font */
	monoFont, err = opentype.NewFace(mono, &opentype.FaceOptions{
		Size:    8,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	monoFontH = getFontHeight(monoFont)

	/* Logo font */
	logoFont, err = opentype.NewFace(logo, &opentype.FaceOptions{
		Size:    70,
		DPI:     fontDPI,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	logoFontH = getFontHeight(logoFont)
}

const sizingText = "!@#$%^&*()_+-=[]{}|;':,.<>?`~qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

func getFontHeight(font font.Face) int {
	defer reportPanic("getFontHeight")
	tRect := text.BoundString(font, sizingText)
	return tRect.Dy()
}
