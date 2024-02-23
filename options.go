package main

import (
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
	optionWindowButtons []image.Rectangle
	settingItems        []settingType
	updateWindowButtons []image.Rectangle
)

type settingType struct {
	ConfigName string
	Text       string `json:"-"`

	TextPosX   int             `json:"-"`
	TextPosY   int             `json:"-"`
	TextBounds image.Rectangle `json:"-"`
	Rect       image.Rectangle `json:"-"`

	Enabled     bool
	WASMExclude bool

	action  func(item int) `json:"-"`
	NoCheck bool           `json:"-"`
}

func init() {
	defer reportPanic("options init")
	settingItems = []settingType{
		{ConfigName: "VSYNC", Text: "Limit FPS (VSYNC)", action: toggleVsync, Enabled: true},
		{ConfigName: "FULLSCREEN", Text: "Full Screen", action: toggleFullscreen},
		{ConfigName: "MAGNIFY", Text: "Magnify UI", action: toggleMagnify},
		{ConfigName: "UNCAP-UPS", Text: "Uncap UPS", action: toggleUPSCap, WASMExclude: true},
		{ConfigName: "DEBUG", Text: "Debug mode", action: toggleDebug, WASMExclude: true},
		{Text: "Load test map", action: toggleTestMap, WASMExclude: true},
		{ConfigName: "FREEDOM-UNITS", Text: "US customary units", action: toggleUnits},
		{ConfigName: "HYPERTHREAD", Text: "Use hyper-threading", action: toggleHyper, WASMExclude: true},
		{ConfigName: "DEBUG-TEXT", Text: "Debug info-text", action: toggleInfoLine},
		{ConfigName: "AUTOSAVE", Text: "Autosave (5m)", action: toggleAutosave, Enabled: true, WASMExclude: true},
		{Text: "Quit game", action: quitGame, NoCheck: true, WASMExclude: true},
	}
}

/* Load user options settings from disk */
func loadOptions() bool {
	defer reportPanic("loadOptions")
	if wasmMode {
		return false
	}

	var tempSettings []settingType

	file, err := os.ReadFile(settingsFile)

	if file != nil && err == nil {

		err := json.Unmarshal([]byte(file), &tempSettings)
		if err != nil {
			doLog(true, "loadOptions: Unmarshal failure")
			doLog(true, err.Error())
			return false
		}
	} else {
		doLog(true, "loadOptions: ReadFile failure")
		return false
	}

	doLog(true, "Settings loaded.")

	for setPos, wSetting := range settingItems {
		for _, fSetting := range tempSettings {
			if wSetting.ConfigName == fSetting.ConfigName {
				if fSetting.Enabled != wSetting.Enabled {
					settingItems[setPos].action(setPos)
				}
			}
		}
	}
	return true
}

/* Save user options settings to disk */
func saveOptions() {
	defer reportPanic("saveOptions")
	if wasmMode {
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

	outBuf := new(bytes.Buffer)
	enc := json.NewEncoder(outBuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(&tempSettings); err != nil {
		doLog(true, "saveOptions: enc.Encode failure")
		return
	}

	os.Mkdir("data", os.ModePerm)
	_, err := os.Create(tempPath)

	if err != nil {
		doLog(true, "saveOptions: os.Create failure")
		return
	}

	err = os.WriteFile(tempPath, outBuf.Bytes(), 0666)

	if err != nil {
		doLog(true, "saveOptions: WriteFile failure")
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		doLog(true, "Couldn't rename settings file.")
		return
	}

	doLog(true, "Settings saved.")
}

/* Toggle the debug bottom-screen text */
func toggleInfoLine(item int) {
	defer reportPanic("toggleInfoLine")
	if infoLine {
		infoLine = false
		settingItems[item].Enabled = false
	} else {
		infoLine = true
		settingItems[item].Enabled = true
	}
}

/* Toggle the use of hyper-threading */
func toggleHyper(item int) {
	defer reportPanic("toggleHyper")
	if useHyper {
		useHyper = false
		settingItems[item].Enabled = false
		detectCPUs(false)
	} else {
		useHyper = true
		settingItems[item].Enabled = true
		detectCPUs(true)
	}
}

/* Close game */
func quitGame(item int) {
	go func() {
		gameRunning = false
		chatDetailed("Game closing...", ColorRed, time.Second*10)
		time.Sleep(time.Second * 2)
		//pprof.StopCPUProfile()
		os.Exit(0)
	}()
}

/* Toggle units */
func toggleUnits(item int) {
	defer reportPanic("toggleUnits")
	if usUnits {
		usUnits = false
		settingItems[item].Enabled = false
	} else {
		usUnits = true
		settingItems[item].Enabled = true
	}
}

/* Toggle test map */
func toggleTestMap(item int) {
	defer reportPanic("toggleTestMap")
	gameRunning = false
	if loadTest {
		loadTest = false
		settingItems[item].Enabled = false
		buf := "Clearing map."
		chatDetailed(buf, ColorOrange, time.Second*10)
	} else {
		loadTest = true
		settingItems[item].Enabled = true
		buf := "Loading test map..."
		chatDetailed(buf, ColorOrange, time.Second*30)
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)

	mapGenerated.Store(false)
	playerReady.Store(0)

	mapLoadPercent = 0
	time.Sleep(time.Millisecond * 10)
	go func() {
		time.Sleep(time.Millisecond * 10)
		makeMap()
		time.Sleep(time.Millisecond * 10)
		startGame()
	}()
}

/* Toggle UPS cap */
func toggleUPSCap(item int) {
	defer reportPanic("toggleUPSCap")
	if upsBench {
		upsBench = false
		settingItems[item].Enabled = false
	} else {
		upsBench = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)
}

/* Toggle full-screen */
func toggleFullscreen(item int) {
	defer reportPanic("toggleFullscreen")
	if ebiten.IsFullscreen() {
		ebiten.SetFullscreen(false)
		settingItems[item].Enabled = false
	} else {
		ebiten.SetFullscreen(true)
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)
}

/* Toggle UI magnification */
func toggleMagnify(item int) {
	defer reportPanic("toggleMagnify")
	if magnify {
		magnify = false
		settingItems[item].Enabled = false
	} else {
		magnify = true
		settingItems[item].Enabled = true
	}

	handleResize(int(ScreenWidth), int(ScreenHeight))

	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)
}

/* Toggle debug mode */
func toggleDebug(item int) {
	defer reportPanic("toggleDebug")
	if debugMode {
		debugMode = false
		settingItems[item].Enabled = false
	} else {
		debugMode = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)
}

/* Toggle autosave */
func toggleAutosave(item int) {
	defer reportPanic("toggleDebug")
	if autoSave {
		autoSave = false
		settingItems[item].Enabled = false
	} else {
		autoSave = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)
}

func toggleVsync(item int) {
	defer reportPanic("toggleVsync")
	if vSync {
		vSync = false
		settingItems[item].Enabled = false
		ebiten.SetVsyncEnabled(false)
	} else {
		vSync = true
		settingItems[item].Enabled = true
		ebiten.SetVsyncEnabled(true)
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		BoolToOnOff(settingItems[item].Enabled))
	chatDetailed(buf, ColorOrange, time.Second*5)
}
