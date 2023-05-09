package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
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
	pngData, err := f.Open(gfxDir + "icon.png")
	if err != nil {
		fmt.Println("Game icon file is missing...")
		return
	}
	m, _, err := image.Decode(pngData)
	if err != nil {
		fmt.Println("Game icon file is invalid...")
		return
	}
	ebiten.SetWindowIcon([]image.Image{m})
}

func getFont(name string) []byte {
	data, err := f.ReadFile(gfxDir + "fonts/" + name)
	if err != nil {
		log.Fatal(err)
	}
	return data

}

func getSpriteImage(name string, unmanaged bool) (*ebiten.Image, error) {

	if cLoadEmbedSprites {
		gpng, err := f.Open(gfxDir + name)
		if err != nil {
			//DoLog(true, "GetSpriteImage: Embedded: %v", err)
			return nil, err
		}

		m, _, err := image.Decode(gpng)
		if err != nil {
			doLog(true, "GetSpriteImage: Embedded: %v", err)
			return nil, err
		}
		var img *ebiten.Image
		if unmanaged {
			img = ebiten.NewImageFromImageWithOptions(m, &ebiten.NewImageFromImageOptions{Unmanaged: true})
		} else {
			img = ebiten.NewImageFromImage(m)
		}
		return img, nil

	} else {
		img, _, err := ebitenutil.NewImageFromFile(dataDir + gfxDir + name)
		if err != nil {
			doLog(true, "GetSpriteImage: File: %v", err)
		}
		return img, err
	}
}

func getText(name string) (string, error) {
	file, err := f.Open(txtDir + name + ".txt")
	if err != nil {
		doLog(true, "GetText: %v", err)
		return "GetText: File: " + name + " not found in embed.", err
	}

	txt, err := io.ReadAll(file)
	if err != nil {
		doLog(true, "GetText: %v", err)
		return "Error: Failed read: " + name, err
	}

	if len(txt) > 0 {
		doLog(true, "GetText: %v", name)
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

func loadSecrets() bool {
	sMutex.Lock()
	defer sMutex.Unlock()

	file, err := f.Open(sFile)
	if err != nil {
		doLog(true, "%v", err)
		return false
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		doLog(true, "%v", err)
		return false
	}

	err = json.Unmarshal(bytes, &Secrets)
	if err != nil {
		doLog(true, "%v", err)
		return false
	}

	return true
}

const emberSavePath = dataDir + "saves/startup.zip"

func loadEmbedSave() []byte {

	file, err := f.Open(emberSavePath)
	if err != nil {
		doLog(true, "Embedded startup save not found: %v", err)
		return nil
	}

	readData, err := io.ReadAll(file)
	if err != nil {
		doLog(true, "Unable to read embedded save game, startup.zip: %v", err)
		return nil
	}

	return readData
}
