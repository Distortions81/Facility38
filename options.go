package main

import (
	"GameTest/cwlog"
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	padding   = 16
	itemsPad  = 16
	linePad   = 10
	spritePad = 32

	TYPE_BOOL = 0
	TYPE_INT  = 1
	TYPE_TEXT = 2

	defaultWindowWidth  = 300
	defaultWindowHeight = 400

	settingsFile = "data/settings.json"
)

var (
	bg          *ebiten.Image
	halfSWidth  = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)
	textHeight  = 16
	windowSizeW = defaultWindowWidth
	windowSizeH = defaultWindowHeight
	halfWindowW = windowSizeW / 2
	halfWindowH = windowSizeH / 2

	buttons      []image.Rectangle
	settingItems []settingType

	closeBoxPos  world.XY
	closeBoxSize world.XY
)

type settingType struct {
	ConfigName string
	Text       string `json:"-"`

	TextPosX   int             `json:"-"`
	TextPosY   int             `json:"-"`
	TextBounds image.Rectangle `json:"-"`
	Rect       image.Rectangle `json:"-"`

	Action  func(item int) `json:"-"`
	Enabled bool
	NoCheck bool `json:"-"`
}

func init() {
	bg = ebiten.NewImage(1, 1)

	bgcolor := color.RGBA{R: 0, G: 0, B: 0, A: 170}
	bg.Fill(bgcolor)

	settingItems = []settingType{
		{ConfigName: "VSYNC", Text: "Limit FPS (VSYNC)", Action: toggleVsync, Enabled: true},
		{ConfigName: "FULLSCREEN", Text: "Full Screen", Action: toggleFullscreen},
		{ConfigName: "UNCAP-FPS", Text: "Uncap UPS", Action: toggleUPSCap},
		{ConfigName: "DEBUG", Text: "Debug mode", Action: toggleDebug},
		{Text: "Load test map", Action: toggleTestMap},
		{ConfigName: "FREEDOM-UNITS", Text: "Imperial Units", Action: toggleUnits},
		{ConfigName: "HYPERTHREAD", Text: "Use hyperthreading", Action: toggleHyper},
		{ConfigName: "DEBUG-TEXT", Text: "Debug info-text", Action: toggleInfoLine},
		{Text: "Quit game", Action: quitGame, NoCheck: true},
	}
}

func loadOptions() {

	if gv.WASMMode {
		return
	}

	var tempSettings []settingType

	file, err := os.ReadFile(settingsFile)

	if file != nil && err == nil {

		err := json.Unmarshal([]byte(file), &tempSettings)
		if err != nil {
			cwlog.DoLog(true, "loadOptions: Unmarshal failure")
			cwlog.DoLog(true, err.Error())
		}
	} else {
		cwlog.DoLog(true, "ReadGCfg: ReadFile failure")
	}

	for wpos, wSetting := range settingItems {
		for _, fSetting := range tempSettings {
			if wSetting.ConfigName == fSetting.ConfigName {
				if fSetting.Enabled != wSetting.Enabled {
					settingItems[wpos].Action(wpos)
				}
			}
		}
	}

}

func saveOptions() {

	if gv.WASMMode {
		return
	}

	var tempSettings []settingType
	for _, setting := range settingItems {
		if setting.ConfigName != "" {
			tempSettings = append(tempSettings, settingType{ConfigName: setting.ConfigName, Enabled: setting.Enabled})
		}
	}

	tempPath := settingsFile + ".tmp"
	finalPath := settingsFile

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(&tempSettings); err != nil {
		cwlog.DoLog(true, "saveOptions: enc.Encode failure")
		return
	}

	_, err := os.Create(tempPath)

	if err != nil {
		cwlog.DoLog(true, "saveOptions: os.Create failure")
		return
	}

	err = os.WriteFile(tempPath, outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog(true, "saveOptions: WriteFile failure")
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		cwlog.DoLog(true, "Couldn't rename Gcfg file.")
		return
	}
}

func toggleInfoLine(item int) {
	if world.InfoLine {
		world.InfoLine = false
		settingItems[item].Enabled = false
	} else {
		world.InfoLine = true
		settingItems[item].Enabled = true
	}
}

func toggleHyper(item int) {
	if world.UseHyper {
		world.UseHyper = false
		settingItems[item].Enabled = false
		detectCPUs(false)
	} else {
		world.UseHyper = true
		settingItems[item].Enabled = true
		detectCPUs(true)
	}
}

func quitGame(item int) {
	go func() {
		objects.GameRunning = false
		util.ChatDetailed("Game closing...", world.ColorRed, time.Second*10)

		objects.GameLock.Lock()
		defer objects.GameLock.Unlock()

		time.Sleep(time.Second * 2)
		os.Exit(0)
	}()
}

func toggleUnits(item int) {
	if world.ImperialUnits {
		world.ImperialUnits = false
		settingItems[item].Enabled = false
	} else {
		world.ImperialUnits = true
		settingItems[item].Enabled = true
	}
}

func toggleTestMap(item int) {
	objects.GameRunning = false
	if gv.LoadTest {
		gv.LoadTest = false
		settingItems[item].Enabled = false
		buf := "Clearing map."
		util.ChatDetailed(buf, world.ColorOrange, time.Second*10)
	} else {
		gv.LoadTest = true
		settingItems[item].Enabled = true
		buf := "Loading test map..."
		util.ChatDetailed(buf, world.ColorOrange, time.Second*30)
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)

	world.MapGenerated.Store(false)
	world.MapLoadPercent = 0
	time.Sleep(time.Nanosecond)
	go func() {
		MakeMap(gv.LoadTest)
		time.Sleep(time.Nanosecond)
		startGame()
	}()
}

func toggleUPSCap(item int) {
	if gv.UPSBench {
		gv.UPSBench = false
		settingItems[item].Enabled = false
	} else {
		gv.UPSBench = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func toggleFullscreen(item int) {
	if ebiten.IsFullscreen() {
		ebiten.SetFullscreen(false)
		settingItems[item].Enabled = false
	} else {
		ebiten.SetFullscreen(true)
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func toggleDebug(item int) {
	if gv.Debug {
		gv.Debug = false
		settingItems[item].Enabled = false
	} else {
		gv.Debug = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func toggleVsync(item int) {
	if world.Vsync {
		world.Vsync = false
		settingItems[item].Enabled = false
		ebiten.SetVsyncEnabled(false)
	} else {
		world.Vsync = true
		settingItems[item].Enabled = true
		ebiten.SetVsyncEnabled(true)
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func setupOptionsMenu() {

	loadOptions()

	if world.BootFont == nil || !world.SpritesLoaded.Load() {
		return
	}

	halfSWidth = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)

	var newVal float32 = 1280.0 / float32(world.ScreenWidth)
	if newVal < 0.1 {
		newVal = 0.1
	} else if newVal > 4 {
		newVal = 4
	}

	windowSizeW = int(defaultWindowWidth / newVal)
	windowSizeH = int(defaultWindowHeight / newVal)

	/* Generate base values */
	halfWindowW = windowSizeW / 2
	halfWindowH = windowSizeH / 2

	base := text.BoundString(world.BootFont, "abcdefghijklmnopqrstuvwxyz!.0123456789")
	textHeight = base.Dy() + linePad
	buttons = []image.Rectangle{}

	img := objects.WorldOverlays[8].Images.Main

	closeBoxPos.X = uint16(halfSWidth + halfWindowW - img.Bounds().Dx() - padding)
	closeBoxPos.Y = uint16(halfSHeight - halfWindowH + padding)
	closeBoxSize.X = uint16(img.Bounds().Dx())
	closeBoxSize.Y = uint16(img.Bounds().Dy())

	/* Loop all settings */
	for i, item := range settingItems {
		/* Get text bounds */
		tbound := text.BoundString(world.BootFont, item.Text)
		settingItems[i].TextBounds = tbound

		/* Place line */
		var linePosX int = (halfSWidth) - halfWindowW + padding
		var linePosY int = (halfSHeight - halfWindowH) + textHeight*(i+2) +
			(linePad * (i + 2)) +
			itemsPad
		settingItems[i].TextPosX = linePosX
		settingItems[i].TextPosY = linePosY

		/* Generate button */
		button := image.Rectangle{}
		button.Min.X = linePosX
		button.Max.X = (halfSWidth + halfWindowW) - padding

		button.Min.Y = linePosY - tbound.Dy()
		button.Max.Y = linePosY + spritePad/2
		buttons = append(buttons, button)
	}
}

func drawSettings(screen *ebiten.Image) {
	halfSWidth = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)
	op := &ebiten.DrawImageOptions{}

	/* Draw window bg */
	op.GeoM.Scale(float64(windowSizeW), float64(windowSizeH))
	op.GeoM.Translate(float64(halfSWidth-halfWindowW), float64(halfSHeight-halfWindowH))
	screen.DrawImage(bg, op)

	/* Draw title */
	txt := "Options:"
	font := world.BootFont
	rect := text.BoundString(font, txt)
	textHeight = rect.Dy() + linePad
	text.Draw(screen, txt, font,
		int(halfSWidth)-(rect.Dx()/2),
		(halfSHeight-halfWindowH)+rect.Dy()+padding,
		world.ColorWhite)

	/* Close box */
	op.GeoM.Reset()
	img := objects.WorldOverlays[8].Images.Main

	op.GeoM.Translate(float64(closeBoxPos.X), float64(closeBoxPos.Y))
	screen.DrawImage(img, op)

	/* Draw items */
	for i, item := range settingItems {
		b := buttons[i]

		/* Text */
		if !item.NoCheck {
			txt = fmt.Sprintf("%v: %v", item.Text, util.BoolToOnOff(item.Enabled))
		} else {
			txt = item.Text
		}

		/* Draw text */
		itemColor := world.ColorWhite
		/* Detect hover, change color */
		mx, my := ebiten.CursorPosition()
		if util.PosWithinRect(world.XY{X: uint16(mx), Y: uint16(my)}, b, 1) {
			itemColor = world.ColorAqua
		}

		/*
			if gv.Debug {
				ebitenutil.DrawRect(screen,
					float64(b.Min.X+((b.Max.X-b.Min.X)/2)-(b.Dx()/2)),
					float64(b.Min.Y+((b.Max.Y-b.Min.Y)/2)-(b.Dy()/2)),
					float64(b.Dx()),
					float64(b.Dy()),
					color.NRGBA{R: 255, G: 0, B: 0, A: 64})
			}
		*/
		text.Draw(screen, txt, font, item.TextPosX, item.TextPosY, itemColor)

		if !item.NoCheck {
			/* Get checkmark image */
			op.GeoM.Reset()
			var check *ebiten.Image
			if item.Enabled {
				check = objects.WorldOverlays[6].Images.Main
			} else {
				check = objects.WorldOverlays[7].Images.Main
			}
			/* Draw checkmark */
			op.GeoM.Translate(
				float64(halfSWidth+halfWindowW-check.Bounds().Dx()-padding),
				float64(item.TextPosY)-float64((check.Bounds().Dy())/2))
			screen.DrawImage(check, op)
		}
	}
}

func handleSettings() bool {
	mx, my := ebiten.CursorPosition()

	for i, item := range settingItems {
		b := buttons[i]
		if util.PosWithinRect(world.XY{X: uint16(mx), Y: uint16(my)}, b, 1) {
			item.Action(i)
			saveOptions()
			gMouseHeld = false
			return true
		}
	}

	/* Close box */
	if mx <= int(closeBoxPos.X+(closeBoxSize.X)) &&
		mx >= int(closeBoxPos.X-(closeBoxSize.X)) &&

		my <= int(closeBoxPos.Y+(closeBoxSize.Y)) &&
		my >= int(closeBoxPos.Y-(closeBoxSize.Y)) {

		world.OptionsOpen = false
		gMouseHeld = false
		return true
	}

	return false
}
