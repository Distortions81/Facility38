package data

import (
	"embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed gfx/ui/save.png
//go:embed gfx/ui/load.png

//go:embed gfx/world-obj/basic-miner.png
//go:embed gfx/world-obj/basic-smelter.png
//go:embed gfx/world-obj/iron-rod-caster.png
//go:embed gfx/world-obj/basic-belt.png
//go:embed gfx/world-obj/basic-belt-vert.png
//go:embed gfx/belt-obj/iron-ore.png
//go:embed gfx/world-obj/basic-box.png
//go:embed gfx/world-obj/steam-engine.png
//go:embed gfx/world-obj/basic-boiler.png

//go:embed gfx/overlays/arrow-north.png
//go:embed gfx/overlays/arrow-south.png
//go:embed gfx/overlays/arrow-east.png
//go:embed gfx/overlays/arrow-west.png

//go:embed gfx/belt-obj/coal-ore.png

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
