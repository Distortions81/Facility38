package main

import (
	"Facility38/cwlog"
	"Facility38/gv"
	"Facility38/objects"
	"Facility38/util"
	"Facility38/world"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	TYPE_BOOL = 0
	TYPE_INT  = 1
	TYPE_TEXT = 2

	settingsFile = "data/settings.json"
)

var (
	textHeight = 16

	buttons      []image.Rectangle
	settingItems []settingType
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
		cwlog.DoLog(true, "loadOptions: ReadFile failure")
		return
	}

	cwlog.DoLog(true, "Settings loaded.")

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
		cwlog.DoLog(true, "Couldn't rename settings file.")
		return
	}

	cwlog.DoLog(true, "Settings saved.")
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

func handleSettings() bool {

	for i, item := range settingItems {
		b := buttons[i]
		if util.PosWithinRect(world.XY{X: uint16(MouseX), Y: uint16(MouseY)}, b, 1) {
			item.Action(i)
			saveOptions()
			gMouseHeld = false
			return true
		}
	}

	return false
}
