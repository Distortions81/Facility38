package data

import (
	"GameTest/consts"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const cLoadEmbedSprites = true

//go:embed txt gfx

var f embed.FS

func GetSpriteImage(name string) (*ebiten.Image, error) {

	if cLoadEmbedSprites {
		gpng, err := f.Open(consts.GfxDir + name)
		if err != nil {
			fmt.Println("GetSpriteImage: Embeded:", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			fmt.Println("GetSpriteImage: Embeded:", err)
			return nil, err
		}
		img := ebiten.NewImageFromImage(m)
		return img, nil

	} else {
		img, _, err := ebitenutil.NewImageFromFile(consts.DataDir + consts.GfxDir + name)
		if err != nil {
			fmt.Println("GetSpriteImage: File:", err)
		}
		return img, err
	}
}

func GetText(name string) (string, error) {
	file, err := f.Open(consts.TxtDir + name + ".txt")
	if err != nil {
		fmt.Println("embed:", err)
		return "GetText: File: " + name + " not found in embed.", err
	}

	txt, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("embed:", err)
		return "Error: Failed read: " + name, err
	}

	if len(txt) > 0 {
		fmt.Println("Loaded text:", name)
		return string(txt), nil
	} else {
		return "Error: length 0!", err
	}

}
