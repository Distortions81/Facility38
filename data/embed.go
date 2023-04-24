package data

import (
	"Facility38/cwlog"
	"Facility38/def"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	//go:embed txt gfx shaders
	f embed.FS

	//PixelateShader *ebiten.Shader
)

const cLoadEmbedSprites = true

/*
func init() {
	var err error
	var shaderProgram []byte
	shaderProgram, err = f.ReadFile(def.ShadersDir + "pixelate.kage")
	if err != nil {
		log.Fatal("Error reading shaders.")
		return
	}

	PixelateShader, err = ebiten.NewShader(shaderProgram)
	if err != nil {
		log.Fatal("Error compiling shaders.")
		return
	}
}
*/

func init() {
	gpng, err := f.Open(def.GfxDir + "icon.png")
	if err != nil {
		fmt.Println("Game icon file is missing...")
		return
	}
	m, _, err := image.Decode(gpng)
	if err != nil {
		fmt.Println("Game icon file is invalid...")
		return
	}
	ebiten.SetWindowIcon([]image.Image{m})
}

func GetFont(name string) []byte {
	data, err := f.ReadFile(def.GfxDir + "fonts/" + name)
	if err != nil {
		log.Fatal(err)
	}
	return data

}

func GetSpriteImage(name string, unmananged bool) (*ebiten.Image, error) {

	if cLoadEmbedSprites {
		gpng, err := f.Open(def.GfxDir + name)
		if err != nil {
			//cwlog.DoLog(true, "GetSpriteImage: Embedded: %v", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			cwlog.DoLog(true, "GetSpriteImage: Embedded: %v", err)
			return nil, err
		}
		var img *ebiten.Image
		if unmananged {
			img = ebiten.NewImageFromImageWithOptions(m, &ebiten.NewImageFromImageOptions{Unmanaged: true})
		} else {
			img = ebiten.NewImageFromImage(m)
		}
		return img, nil

	} else {
		img, _, err := ebitenutil.NewImageFromFile(def.DataDir + def.GfxDir + name)
		if err != nil {
			cwlog.DoLog(true, "GetSpriteImage: File: %v", err)
		}
		return img, err
	}
}

func GetText(name string) (string, error) {
	file, err := f.Open(def.TxtDir + name + ".txt")
	if err != nil {
		cwlog.DoLog(true, "GetText: %v", err)
		return "GetText: File: " + name + " not found in embed.", err
	}

	txt, err := io.ReadAll(file)
	if err != nil {
		cwlog.DoLog(true, "GetText: %v", err)
		return "Error: Failed read: " + name, err
	}

	if len(txt) > 0 {
		cwlog.DoLog(true, "GetText: %v", name)
		return strings.ReplaceAll(string(txt), "\r", ""), nil
	} else {
		return "Error: length 0!", err
	}

}
