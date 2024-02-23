package main

import (
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	/* UI state */
	gMouseHeld       bool
	gMiddleMouseHeld bool
	gRightMouseHeld  bool
	gShiftPressed    bool
	gClickCaptured   bool
	gCameraDrag      bool
	gWindowDrag      *windowData

	/* Last object we performed an action on */
	gLastActionPosition XY

	/* WASM weirdness kludge */
	lastScroll time.Time

	MouseX int
	MouseY int

	lastMouseX int
	LastMouseY int
)

func init() {
	defer reportPanic("ui init")
	lastScroll = time.Now()
}

/* Input interface handler */
func (g *Game) Update() error {
	defer reportPanic("Update")

	if audioPlayer != nil && !audioPlayer.IsPlaying() {
		audioPlayer.Rewind()
		audioPlayer.SetVolume(0.20)
		audioPlayer.Play()
		doLog(true, "Playing music...")
	}

	/* Reset click 'captured' state */
	gClickCaptured = false

	/* Ignore if game not focused */
	if !ebiten.IsFocused() {
		return nil
	}

	/* Save mouse coords */
	lastMouseX = MouseX
	LastMouseY = MouseY

	/* Clamp to window */
	MouseX, MouseY = ebiten.CursorPosition()
	if MouseX < 0 || MouseX > int(ScreenWidth) ||
		MouseY < 0 || MouseY > int(ScreenHeight) {
		MouseX = lastMouseX
		MouseY = LastMouseY

		/* Stop dragging window if we go off-screen */
		gWindowDrag = nil
		gClickCaptured = true //Eat the click
	}

	var keys []ebiten.Key
	/* Game start screen */
	if (playerReady.Load() == 0 || !mapGenerated.Load()) &&
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.AppendPressedKeys(keys) != nil) {
		playerReady.Store(1)
		return nil
	}

	/* Server says no input for you */
	if !authorized.Load() {
		return nil
	}

	/* Detect inputs */
	getMouseClicks()
	getMiddleMouseClicks()
	getRightMouseClicks()
	getShiftToggle()
	getToolbarKeyPress()

	handleQuit()
	zoomHandle()

	/* Check if we clicked within a window */
	gClickCaptured = collisionWindowsCheck(XYs{X: int32(MouseX), Y: int32(MouseY)})

	/* Handle window drag */
	if gWindowDrag != nil {
		gWindowDrag.position = XYs{X: int32(MouseX) - gWindowDrag.dragPos.X, Y: int32(MouseY) - gWindowDrag.dragPos.Y}
		gClickCaptured = true
	}

	/* If we aren't moving a window, or clicking in one, click goes to game world */
	if gWindowDrag == nil && !gClickCaptured {
		createWorldObjects()
		moveCamera()
		rotateWorldObjects()
	}

	/* Update screen position calculations */
	calcScreenCamera()

	/* Get mouse position on world */
	worldMouseX = (float32(MouseX)/zoomScale + (cameraX - (float32(ScreenWidth)/2.0)/zoomScale))
	worldMouseY = (float32(MouseY)/zoomScale + (cameraY - (float32(ScreenHeight)/2.0)/zoomScale))
	return nil
}

/* Toolbar shortcut keys */
func getToolbarKeyPress() {
	defer reportPanic("getToolbarKeypress")
	for _, item := range uiObjs {
		if inpututil.IsKeyJustPressed(item.qKey) {
			item.toolbarAction()
		}
	}
}

/* Quit if alt-f4 or ESC are pressed */
func handleQuit() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) && ebiten.IsKeyPressed(ebiten.KeyAlt) {
		quitGame(0)
	}
}

/* Record shift state */
func getShiftToggle() {
	defer reportPanic("getShiftToggle")
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		gShiftPressed = true
	} else if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		gShiftPressed = false
	}
}

/* Handle clicks that end up within the toolbar */
func handleToolbar(rotate bool) bool {
	defer reportPanic("handleToolbar")
	iconSize := float32(uiScale * toolBarIconSize)
	spacing := float32(toolBarIconSize / toolBarSpaceRatio)

	tbLength := float32((toolbarMax * int(iconSize+spacing)))

	fmx := float32(MouseX)
	fmy := float32(MouseY)

	/* If the click isn't off the right of the toolbar */
	if fmx <= tbLength {
		/* If the click isn't below the toolbar */
		if fmy <= iconSize {

			tbItem := int(fmx / float32(iconSize+spacing))
			len := len(toolbarItems) - 1
			if tbItem > len {
				tbItem = len
			} else if tbItem < 0 {
				tbItem = 0
			}
			item := toolbarItems[tbItem].oType

			/* Actions */
			if item.toolbarAction != nil && !rotate {
				item.toolbarAction()
			} else {
				/* Not a click, check for rotation */
				if rotate && item != nil {
					dir := item.direction
					if gShiftPressed {
						dir = RotCCW(dir)
					} else {
						dir = RotCW(dir)
					}
					item.direction = dir

					/* Deselect */
				} else if selectedItemType == toolbarItems[tbItem].oType.typeI {
					selectedItemType = maxItemType

				} else {
					/* Select */
					selectedItemType = toolbarItems[tbItem].oType.typeI

				}
			}

			drawToolbar(true, false, tbItem)

			/* Eat this mouse event */
			gMouseHeld = false
			gClickCaptured = true
			return true
		}
	}
	return false
}

/* Handle scroll wheel and +- keys */
var zoomLock sync.Mutex

func zoomHandle() {
	defer reportPanic("zoomHandle")

	zoomLock.Lock()
	defer zoomLock.Unlock()

	/* Zoom in or out with keyboard */
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		zoomScale = zoomScale * 2
		limitZoom()
		visDataDirty.Store(true)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		zoomScale = zoomScale / 2
		limitZoom()
		visDataDirty.Store(true)
	}

	/* Mouse scroll zoom */
	_, fsy := ebiten.Wheel()

	/* WASM weirdness kludge */
	if wasmMode && (fsy > 0.0001 || fsy < -0.0001) {
		if time.Since(lastScroll) < (time.Millisecond * 200) {
			visDataDirty.Store(true)
			return
		}
	}
	lastScroll = time.Now()

	if fsy > 0 {
		/* Zoom in with scroll wheel */
		zoomScale = zoomScale * 2

		/* Center world on mouse */
		if limitZoom() {
			cameraX = worldMouseX
			cameraY = worldMouseY
		}
		visDataDirty.Store(true)
	} else if fsy < 0 {
		/* Zoom out */
		zoomScale = zoomScale / 2
		limitZoom()
		visDataDirty.Store(true)
	}

}

/* Clamp zoom to a range */
func limitZoom() bool {
	defer reportPanic("limitZoom")
	if zoomScale < 1 {
		zoomScale = 1
		visDataDirty.Store(true)
		return false
	} else if zoomScale > 256 {
		zoomScale = 256
		visDataDirty.Store(true)
		return false
	}

	return true
}

/* Record mouse clicks, send clicks to toolbar */
func getMouseClicks() {
	defer reportPanic("getMouseClicks")
	/* Mouse clicks */
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		gMouseHeld = false

		/* Stop dragging window */
		gWindowDrag = nil

		gLastActionPosition = XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		gMouseHeld = true
		gLastActionPosition.X = 0
		gLastActionPosition.Y = 0
		handleToolbar(false)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		handleToolbar(true)
	}
}

/* Look for clicks in window, create or destroy objects */
func createWorldObjects() {
	defer reportPanic("createWorldObjects")
	if gClickCaptured {
		return
	}

	/* Is mouse held */
	if !gMouseHeld && !gRightMouseHeld {
		return
	}

	/* Has the click already been captured? */
	if gClickCaptured {
		return
	}
	pos := FloatXYToPosition(worldMouseX, worldMouseY)

	/* Is this a new position? */
	if pos == gLastActionPosition {
		return
	}
	gLastActionPosition = pos

	chunk := GetChunk(pos)
	b := GetObj(pos, chunk)

	/* Left mouse, and tile is empty */
	if gMouseHeld {
		/* Nothing selected exit */
		if selectedItemType == maxItemType {
			return
		}
		obj := worldObjs[selectedItemType]
		dir := obj.direction

		objQueueAdd(nil, selectedItemType, pos, false, dir)

		/* Else if tile is not empty and RightMouse is held */
	} else if b != nil && b.obj != nil && gRightMouseHeld {
		objQueueAdd(b.obj, selectedItemType, b.obj.Pos, true, 0)
	}
}

/* Right-click drag or WASD movement, shift run */
var lastUpdate time.Time

/* Move camera, based on wall time */
func moveCamera() {
	defer reportPanic("moveCamera")

	var startBase float64 = moveSpeed
	if gShiftPressed {
		startBase = runSpeed
	}

	/* Adjust speed based on high-percision TPS */
	tps := 1000000000.0 / float64(time.Since(lastUpdate).Nanoseconds())
	lastUpdate = time.Now()
	base := startBase / (float64(tps / 60.0))

	/* Base speed on zoom level */
	speed := float32(base / (float64(zoomScale) / 4.0))

	/* WASD keys */
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		cameraY -= speed
		visDataDirty.Store(true)

	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		cameraX -= speed
		visDataDirty.Store(true)

	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		cameraY += speed
		visDataDirty.Store(true)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		cameraX += speed
		visDataDirty.Store(true)
	}

	/* Middle-Mouse click-drag */
	if gMiddleMouseHeld {

		if !gCameraDrag {
			gCameraDrag = true
		}

		cameraX = cameraX + (float32(lastMouseX-MouseX) / zoomScale)
		cameraY = cameraY + (float32(LastMouseY-MouseY) / zoomScale)
		visDataDirty.Store(true)
		lastMouseX = MouseX
		LastMouseY = MouseY

		/* Don't let camera go beyond a reasonable point */
		if cameraX > float32(xyMax) {
			cameraX = float32(xyMax)
		} else if cameraX < xyMin {
			cameraX = xyMin
		}
		if cameraY > float32(xyMax) {
			cameraY = float32(xyMax)
		} else if cameraY < xyMin {
			cameraY = xyMin
		}
	} else {
		gCameraDrag = false
	}
}

/* Detect and record right click state */
func getRightMouseClicks() {
	defer reportPanic("getRightMouseClicks")
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		gRightMouseHeld = false
		gLastActionPosition = XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		gRightMouseHeld = true
	}
}

/* Detect middle mouse click */
func getMiddleMouseClicks() {
	defer reportPanic("getMiddleMouseClicks")
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = false
		gLastActionPosition = XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = true
	}
}

/* Detect R key */
func rotateWorldObjects() {
	defer reportPanic("rotateWorldObjects")
	/* Rotate object */
	if !gClickCaptured && inpututil.IsKeyJustPressed(ebiten.KeyR) {

		pos := FloatXYToPosition(worldMouseX, worldMouseY)

		chunk := GetChunk(pos)
		/* Valid chunk? */
		if chunk == nil {
			return
		}

		b := GetObj(pos, chunk)
		/* Valid building? */

		if b == nil {
			/* Nothing is selected, exit */
			if selectedItemType == maxItemType {
				return
			}

			for pos := 0; pos < toolbarMax; pos++ {
				if toolbarItems[pos].oType == nil {
					continue
				}
				item := toolbarItems[pos].oType
				if item.typeI != selectedItemType {
					continue
				}
				if !item.rotatable {
					continue
				}

				dir := item.direction
				if gShiftPressed {
					dir = RotCCW(dir)
				} else {
					dir = RotCW(dir)
				}
				item.direction = dir

				drawToolbar(true, false, pos)
			}

			return
		}

		/* Valid object? */
		if b.obj == nil {
			return
		}

		/* Queue up a object rotation */
		rotateListAdd(b, !gShiftPressed, pos)
	}
}
