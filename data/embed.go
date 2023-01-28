package data

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

/*go:embed gfx */
var f embed.FS

func GetSpriteImage(embeded bool, name string) (*ebiten.Image, error) {

	if embeded {
		gpng, err := f.Open(name)
		if err != nil {
			fmt.Println("embed:", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			fmt.Println("embed:", err)
			return nil, err
		}
		img := ebiten.NewImageFromImage(m)
		return img, nil

	} else {
		img, _, err := ebitenutil.NewImageFromFile("data/" + name)
		if err != nil {
			fmt.Println("load:", err)
		}
		return img, err
	}
}
