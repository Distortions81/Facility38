package main

import (
	"Facility38/gv"
	"Facility38/objects"
	"Facility38/util"
	"Facility38/world"
	"os"
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

	/* Last object we performed an action on */
	gLastActionPosition world.XY

	/* WASM weirdness kludge */
	lastScroll time.Time

	UILastMouseX int
	UILastMouseY int
)

func init() {
	lastScroll = time.Now()
}

/* Input interface handler */
func (g *Game) Update() error {

	var keys []ebiten.Key
	/* Game start screen */
	if (!world.PlayerReady.Load() || !world.MapGenerated.Load()) &&
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.AppendPressedKeys(keys) != nil) {
		world.PlayerReady.Store(true)
		return nil
	}
	gClickCaptured = false

	getMouseClicks()
	getMiddleMouseClicks()
	getRightMouseClicks()
	getShiftToggle()
	getToolbarKeypress()

	handleQuit()

	//touchScreenHandle()
	zoomHandle()

	createWorldObjects()
	moveCamera()
	rotateWorldObjects()

	mx, my := ebiten.CursorPosition()
	UILastMouseX = mx
	UILastMouseY = my
	return nil
}

func getToolbarKeypress() {

	for _, item := range objects.UIObjs {
		if inpututil.IsKeyJustPressed(item.QKey) {
			item.ToolbarAction()
		}
	}
}

/* Quit if alt-f4 or ESC are pressed */
func handleQuit() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) && ebiten.IsKeyPressed(ebiten.KeyAlt) {
		objects.GameRunning = false
		util.ChatDetailed("Game closing...", world.ColorRed, time.Second*10)

		objects.GameLock.Lock()
		defer objects.GameLock.Unlock()

		time.Sleep(time.Second * 2)
		os.Exit(0)
	}
}

/* Record shift state */
func getShiftToggle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		gShiftPressed = true
	} else if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		gShiftPressed = false
	}
}

/* Handle clicks that end up within the toolbar */
func handleToolbar(rotate bool) bool {
	uipix := float32((ToolbarMax * int(gv.ToolBarScale+gv.ToolBarSpacing)))

	mx, my := ebiten.CursorPosition()
	fmx := float32(mx)
	fmy := float32(my)

	if fmx <= uipix {
		if fmy <= gv.ToolBarScale {

			ipos := int(fmx / float32(gv.ToolBarScale+gv.ToolBarSpacing))
			len := len(ToolbarItems) - 1
			if ipos > len {
				ipos = len
			} else if ipos < 0 {
				ipos = 0
			}
			item := ToolbarItems[ipos].OType

			DrawToolbar(true, false, ipos)

			/* Actions */
			if item.ToolbarAction != nil && !rotate {
				item.ToolbarAction()
				DrawToolbar(true, false, ipos)
			} else {
				if rotate && item != nil {
					dir := item.Direction
					if gShiftPressed {
						dir = util.RotCCW(dir)
					} else {
						dir = util.RotCW(dir)
					}
					item.Direction = dir
					DrawToolbar(true, false, ipos)
					/* Deselect */
				} else if SelectedItemType == ToolbarItems[ipos].OType.TypeI {
					SelectedItemType = gv.MaxItemType
					DrawToolbar(true, false, ipos)
				} else {
					/* Select */
					SelectedItemType = ToolbarItems[ipos].OType.TypeI
					DrawToolbar(true, false, ipos)
				}
			}
			gMouseHeld = false
			return true
		}
	}
	return false
}

/* Handle scroll wheel and +- keys */
func zoomHandle() {
	/* Mouse scroll zoom */
	_, fsy := ebiten.Wheel()

	/* WASM kludge */
	if gv.WASMMode && (fsy > 0 && fsy < 0) {
		if time.Since(lastScroll) < (time.Millisecond * 200) {
			world.VisDataDirty.Store(true)
			return
		}
	}
	lastScroll = time.Now()

	mx, my := ebiten.CursorPosition()
	fmx := float32(mx)
	fmy := float32(my)

	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		world.ZoomScale = world.ZoomScale * 2
		limitZoom()
		world.VisDataDirty.Store(true)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		world.ZoomScale = world.ZoomScale / 2
		limitZoom()
		world.VisDataDirty.Store(true)
	} else if fsy > 0 {
		world.ZoomScale = world.ZoomScale * 2
		if limitZoom() {
			/* Get mouse position on world */
			worldMouseX := (fmx/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
			worldMouseY := (fmy/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))
			world.CameraX = worldMouseX
			world.CameraY = worldMouseY
		}
		world.VisDataDirty.Store(true)
	} else if fsy < 0 {
		world.ZoomScale = world.ZoomScale / 2
		limitZoom()
		world.VisDataDirty.Store(true)
	}

}

func limitZoom() bool {
	if world.ZoomScale < 1 {
		world.ZoomScale = 1
		world.VisDataDirty.Store(true)
		return false
	} else if world.ZoomScale > 256 {
		world.ZoomScale = 256
		world.VisDataDirty.Store(true)
		return false
	}

	return true
}

/* Record mouse clicks, send clicks to toolbar */
func getMouseClicks() {
	/* Mouse clicks */
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		gMouseHeld = false
		gLastActionPosition = world.XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		gMouseHeld = true
		gLastActionPosition.X = 0
		gLastActionPosition.Y = 0

		if world.OptionsOpen {
			gClickCaptured = handleSettings()
		}
		if !gClickCaptured {
			gClickCaptured = handleToolbar(false)
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		gClickCaptured = handleToolbar(true)
	}
}

/* Look for clicks in window, create or destroy objects */
func createWorldObjects() {
	/* Is mouse held */
	if !gMouseHeld && !gRightMouseHeld {
		return
	}

	/* Has the click already been captured? */
	if gClickCaptured {
		return
	}

	/* Settings open? */
	if world.OptionsOpen {
		return
	}

	mx, my := ebiten.CursorPosition()
	fmx := float32(mx)
	fmy := float32(my)

	/* Get mouse position on world */
	worldMouseX := (fmx/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
	worldMouseY := (fmy/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))

	pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

	/* Is this a new position? */
	if pos == gLastActionPosition {
		return
	}
	gLastActionPosition = pos

	chunk := util.GetChunk(pos)
	b := util.GetObj(pos, chunk)

	/* Left mouse, and tile is empty */
	if gMouseHeld {
		/* Nothing selected exit */
		if SelectedItemType == gv.MaxItemType {
			return
		}
		obj := objects.WorldObjs[SelectedItemType]
		dir := obj.Direction

		if gv.WASMMode {
			objects.ObjQueueAdd(nil, SelectedItemType, pos, false, dir)
		} else {
			go objects.ObjQueueAdd(nil, SelectedItemType, pos, false, dir)
		}

		/* Else if tile is not empty and RightMouse is held */
	} else if b != nil && b.Obj != nil && gRightMouseHeld {
		if gv.WASMMode {
			objects.ObjQueueAdd(b.Obj, SelectedItemType, b.Obj.Pos, true, 0)
		} else {
			go objects.ObjQueueAdd(b.Obj, SelectedItemType, b.Obj.Pos, true, 0)
		}
	}
}

/* Right-click drag or WASD movement, shift run */
func moveCamera() {

	var base float32 = gv.MoveSpeed
	if gShiftPressed {
		base = gv.RunSpeed
	}
	speed := base / (world.ZoomScale / 4.0)

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		world.CameraY -= speed
		world.VisDataDirty.Store(true)

	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		world.CameraX -= speed
		world.VisDataDirty.Store(true)

	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		world.CameraY += speed
		world.VisDataDirty.Store(true)

	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		world.CameraX += speed
		world.VisDataDirty.Store(true)

	}

	if gMiddleMouseHeld {
		mx, my := ebiten.CursorPosition()

		if !gCameraDrag {
			gCameraDrag = true
		}

		world.CameraX = world.CameraX + (float32(UILastMouseX-mx) / world.ZoomScale)
		world.CameraY = world.CameraY + (float32(UILastMouseY-my) / world.ZoomScale)
		world.VisDataDirty.Store(true)

		/* Don't let camera go beyond a reasonable point */
		if world.CameraX > float32(gv.XYMax) {
			world.CameraX = float32(gv.XYMax)
		} else if world.CameraX < gv.XYMin {
			world.CameraX = gv.XYMin
		}
		if world.CameraY > float32(gv.XYMax) {
			world.CameraY = float32(gv.XYMax)
		} else if world.CameraY < gv.XYMin {
			world.CameraY = gv.XYMin
		}
	} else {
		gCameraDrag = false
	}
}

/* Detect and record right click state */
func getRightMouseClicks() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		gRightMouseHeld = false
		gLastActionPosition = world.XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		gRightMouseHeld = true
	}
}

func getMiddleMouseClicks() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = false
		gLastActionPosition = world.XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = true
	}
}

func rotateWorldObjects() {
	/* Rotate object */
	if !gClickCaptured && inpututil.IsKeyJustPressed(ebiten.KeyR) {

		/* Get mouse position on world */
		mx, my := ebiten.CursorPosition()
		fmx := float32(mx)
		fmy := float32(my)
		worldMouseX := (fmx/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth/2.0) / world.ZoomScale)))
		worldMouseY := (fmy/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight/2.0))/world.ZoomScale))

		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

		chunk := util.GetChunk(pos)
		/* Valid chunk? */
		if chunk == nil {
			return
		}

		b := util.GetObj(pos, chunk)
		/* Valid building? */
		if b == nil {
			return
		}

		/* Valid object? */
		if b.Obj == nil {
			return
		}

		objects.RotateListAdd(b, !gShiftPressed, pos)
	}
}
