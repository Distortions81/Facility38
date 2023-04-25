package main

import (
	"Facility38/cwlog"
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
		{ConfigName: "MAGNIFY", Text: "Magnifiy UI", Action: toggleMagnify},
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

	if world.WASMMode {
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

	if world.WASMMode {
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
		GameRunning = false
		util.ChatDetailed("Game closing...", world.ColorRed, time.Second*10)

		GameLock.Lock()
		defer GameLock.Unlock()

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
	GameRunning = false
	if world.LoadTest {
		world.LoadTest = false
		settingItems[item].Enabled = false
		buf := "Clearing map."
		util.ChatDetailed(buf, world.ColorOrange, time.Second*10)
	} else {
		world.LoadTest = true
		settingItems[item].Enabled = true
		buf := "Loading test map..."
		util.ChatDetailed(buf, world.ColorOrange, time.Second*30)
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)

	world.MapGenerated.Store(false)
	world.PlayerReady.Store(0)

	world.MapLoadPercent = 0
	time.Sleep(time.Millisecond * 10)
	go func() {
		time.Sleep(time.Millisecond * 10)
		MakeMap(world.LoadTest)
		time.Sleep(time.Millisecond * 10)
		startGame()
	}()
}

func toggleUPSCap(item int) {
	if world.UPSBench {
		world.UPSBench = false
		settingItems[item].Enabled = false
	} else {
		world.UPSBench = true
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

func toggleMagnify(item int) {
	if world.Magnify {
		world.Magnify = false
		settingItems[item].Enabled = false
	} else {
		world.Magnify = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
	ow, oh := ebiten.WindowSize()
	if ow > 0 && oh > 0 {
		handleResize(ow, oh)
	}

}

func toggleDebug(item int) {
	if world.Debug {
		world.Debug = false
		settingItems[item].Enabled = false
	} else {
		world.Debug = true
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
