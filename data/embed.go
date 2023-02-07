package data

import (
	"GameTest/cwlog"
	"GameTest/gv"
	"embed"
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
		gpng, err := f.Open(gv.GfxDir + name)
		if err != nil {
			cwlog.DoLog("GetSpriteImage: Embeded: %v", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			cwlog.DoLog("GetSpriteImage: Embeded: %v", err)
			return nil, err
		}
		img := ebiten.NewImageFromImage(m)
		return img, nil

	} else {
		img, _, err := ebitenutil.NewImageFromFile(gv.DataDir + gv.GfxDir + name)
		if err != nil {
			cwlog.DoLog("GetSpriteImage: File: %v", err)
		}
		return img, err
	}
}

func GetText(name string) (string, error) {
	file, err := f.Open(gv.TxtDir + name + ".txt")
	if err != nil {
		cwlog.DoLog("GetText: %v", err)
		return "GetText: File: " + name + " not found in embed.", err
	}

	txt, err := io.ReadAll(file)
	if err != nil {
		cwlog.DoLog("GetText: %v", err)
		return "Error: Failed read: " + name, err
	}

	if len(txt) > 0 {
		cwlog.DoLog("GetText: %v", name)
		return string(txt), nil
	} else {
		return "Error: length 0!", err
	}

}
