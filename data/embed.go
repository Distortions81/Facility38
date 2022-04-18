package data

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var f embed.FS

func GetSpriteImage(embeded bool, name string) (*ebiten.Image, error) {

	var err error
	if embeded && 1 == 2 {
		gpng, err := f.Open(name)
		if err == nil {

			m, _, err := image.Decode(gpng)
			if err == nil {
				img := ebiten.NewImageFromImage(m)
				return img, nil
			}
		}
	} else {
		img, _, err := ebitenutil.NewImageFromFile("data/" + name)
		if err != nil {
			fmt.Println(err)
		}
		return img, err
	}
	return nil, err
}
