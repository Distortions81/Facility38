package data

import (
	"embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed gfx/icons/save.png
//go:embed gfx/icons/miner.png
//go:embed gfx/icons/furnace.png

var f embed.FS

func GetSpriteImage(embeded bool, name string) (*ebiten.Image, error) {
	if embeded {
		gpng, err := f.Open(name)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		m, _, err := image.Decode(gpng)
		if err == nil {
			img := ebiten.NewImageFromImage(m)
			return img, nil
		}
		return nil, err
	} else {
		img, _, err := ebitenutil.NewImageFromFile(name)
		return img, err
	}

}
