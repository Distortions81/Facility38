package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	//go:embed data
	f embed.FS
)

const cLoadEmbedSprites = true

func init() {
	gpng, err := f.Open(gfxDir + "icon.png")
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
	data, err := f.ReadFile(gfxDir + "fonts/" + name)
	if err != nil {
		log.Fatal(err)
	}
	return data

}

func GetSpriteImage(name string, unmananged bool) (*ebiten.Image, error) {

	if cLoadEmbedSprites {
		gpng, err := f.Open(gfxDir + name)
		if err != nil {
			//DoLog(true, "GetSpriteImage: Embedded: %v", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			DoLog(true, "GetSpriteImage: Embedded: %v", err)
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
		img, _, err := ebitenutil.NewImageFromFile(dataDir + gfxDir + name)
		if err != nil {
			DoLog(true, "GetSpriteImage: File: %v", err)
		}
		return img, err
	}
}

func GetText(name string) (string, error) {
	file, err := f.Open(txtDir + name + ".txt")
	if err != nil {
		DoLog(true, "GetText: %v", err)
		return "GetText: File: " + name + " not found in embed.", err
	}

	txt, err := io.ReadAll(file)
	if err != nil {
		DoLog(true, "GetText: %v", err)
		return "Error: Failed read: " + name, err
	}

	if len(txt) > 0 {
		DoLog(true, "GetText: %v", name)
		return strings.ReplaceAll(string(txt), "\r", ""), nil
	} else {
		return "Error: length 0!", err
	}

}

const sFile = txtDir + "p.json"

var Secrets []secData
var sMutex sync.Mutex

type secData struct {
	P string `json:"p,omitempty"`
	R string `json:"r,omitempty"`
}

func LoadSecrets() bool {
	sMutex.Lock()
	defer sMutex.Unlock()

	file, err := f.Open(sFile)
	if err != nil {
		DoLog(true, "%v", err)
		return false
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		DoLog(true, "%v", err)
		return false
	}

	err = json.Unmarshal(bytes, &Secrets)
	if err != nil {
		DoLog(true, "%v", err)
		return false
	}

	return true
}
