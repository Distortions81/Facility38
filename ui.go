package main

import (
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
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
	gWindowDrag      *WindowData

	/* Last object we performed an action on */
	gLastActionPosition world.XY

	/* WASM weirdness kludge */
	lastScroll time.Time

	MouseX int
	MouseY int

	LastMouseX int
	LastMouseY int
)

func init() {
	defer util.ReportPanic("ui init")
	lastScroll = time.Now()
}

/* Input interface handler */
func (g *Game) Update() error {
	defer util.ReportPanic("Update")
	gClickCaptured = false
	/* Ignore if not focused */
	if !ebiten.IsFocused() {
		return nil
	}

	LastMouseX = MouseX
	LastMouseY = MouseY

	/* Clamp to window */
	MouseX, MouseY = ebiten.CursorPosition()
	if MouseX < 0 || MouseX > int(world.ScreenWidth) ||
		MouseY < 0 || MouseY > int(world.ScreenHeight) {
		MouseX = LastMouseX
		MouseY = LastMouseY

		/* Stop dragging window if we go off-screen */
		gWindowDrag = nil
		gClickCaptured = true
	}

	var keys []ebiten.Key
	/* Game start screen */
	if (world.PlayerReady.Load() == 0 || !world.MapGenerated.Load()) &&
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.AppendPressedKeys(keys) != nil) {
		world.PlayerReady.Store(1)
		return nil
	}

	if !world.Authorized.Load() {
		return nil
	}

	getMouseClicks()
	getMiddleMouseClicks()
	getRightMouseClicks()
	getShiftToggle()
	getToolbarKeypress()

	handleQuit()

	zoomHandle()

	gClickCaptured = CollisionWindowsCheck(world.XYs{X: int32(MouseX), Y: int32(MouseY)})

	/* Handle window drag */
	if gWindowDrag != nil {
		gWindowDrag.Position = world.XYs{X: int32(MouseX) - gWindowDrag.DragPos.X, Y: int32(MouseY) - gWindowDrag.DragPos.Y}
		gClickCaptured = true
	}

	if gWindowDrag == nil && !gClickCaptured {
		createWorldObjects()
		moveCamera()
		rotateWorldObjects()
	}

	calcScreenCamera()
	/* Get mouse position on world */
	WorldMouseX = (float32(MouseX)/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
	WorldMouseY = (float32(MouseY)/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))
	return nil
}

func getToolbarKeypress() {
	defer util.ReportPanic("getToolbarKeypress")
	for _, item := range UIObjs {
		if inpututil.IsKeyJustPressed(item.QKey) {
			item.ToolbarAction()
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
	defer util.ReportPanic("getShiftToggle")
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		gShiftPressed = true
	} else if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		gShiftPressed = false
	}
}

/* Handle clicks that end up within the toolbar */
func handleToolbar(rotate bool) bool {
	defer util.ReportPanic("handleToolbar")
	ToolBarIconSize := float32(world.UIScale * def.ToolBarIconSize)
	ToolBarSpacing := float32(def.ToolBarIconSize / def.ToolBarSpaceRatio)

	uipix := float32((ToolbarMax * int(ToolBarIconSize+ToolBarSpacing)))

	fmx := float32(MouseX)
	fmy := float32(MouseY)

	if fmx <= uipix {
		if fmy <= ToolBarIconSize {

			ipos := int(fmx / float32(ToolBarIconSize+ToolBarSpacing))
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
					SelectedItemType = def.MaxItemType
					DrawToolbar(true, false, ipos)
				} else {
					/* Select */
					SelectedItemType = ToolbarItems[ipos].OType.TypeI
					DrawToolbar(true, false, ipos)
				}
			}
			gMouseHeld = false
			gClickCaptured = true
			return true
		}
	}
	return false
}

/* Handle scroll wheel and +- keys */
func zoomHandle() {
	defer util.ReportPanic("zoomHandle")
	/* Mouse scroll zoom */
	_, fsy := ebiten.Wheel()

	/* WASM kludge */
	if world.WASMMode && (fsy > 0 && fsy < 0) {
		if time.Since(lastScroll) < (time.Millisecond * 200) {
			world.VisDataDirty.Store(true)
			return
		}
	}
	lastScroll = time.Now()

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
			world.CameraX = WorldMouseX
			world.CameraY = WorldMouseY
		}
		world.VisDataDirty.Store(true)
	} else if fsy < 0 {
		world.ZoomScale = world.ZoomScale / 2
		limitZoom()
		world.VisDataDirty.Store(true)
	}

}

func limitZoom() bool {
	defer util.ReportPanic("limitZoom")
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
	defer util.ReportPanic("getMouseClicks")
	/* Mouse clicks */
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		gMouseHeld = false

		/* Stop dragging window */
		gWindowDrag = nil

		gLastActionPosition = world.XY{}
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
	defer util.ReportPanic("createWorldObjects")
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
	pos := util.FloatXYToPosition(WorldMouseX, WorldMouseY)

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
		if SelectedItemType == def.MaxItemType {
			return
		}
		obj := WorldObjs[SelectedItemType]
		dir := obj.Direction

		if world.WASMMode {
			ObjQueueAdd(nil, SelectedItemType, pos, false, dir)
		} else {
			go ObjQueueAdd(nil, SelectedItemType, pos, false, dir)
		}

		/* Else if tile is not empty and RightMouse is held */
	} else if b != nil && b.Obj != nil && gRightMouseHeld {
		if world.WASMMode {
			ObjQueueAdd(b.Obj, SelectedItemType, b.Obj.Pos, true, 0)
		} else {
			go ObjQueueAdd(b.Obj, SelectedItemType, b.Obj.Pos, true, 0)
		}
	}
}

/* Right-click drag or WASD movement, shift run */
var lastUpdate time.Time

func moveCamera() {
	defer util.ReportPanic("moveCamera")
	var startBase float64 = def.MoveSpeed
	if gShiftPressed {
		startBase = def.RunSpeed
	}

	/* Adjust speed based on high-percision TPS */
	tps := 1000000000.0 / float64(time.Since(lastUpdate).Nanoseconds())
	lastUpdate = time.Now()
	base := startBase / (float64(tps / 60.0))

	/* Base speed on zoom level */
	speed := float32(base / (float64(world.ZoomScale) / 4.0))

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

		if !gCameraDrag {
			gCameraDrag = true
		}

		world.CameraX = world.CameraX + (float32(LastMouseX-MouseX) / world.ZoomScale)
		world.CameraY = world.CameraY + (float32(LastMouseY-MouseY) / world.ZoomScale)
		world.VisDataDirty.Store(true)
		LastMouseX = MouseX
		LastMouseY = MouseY

		/* Don't let camera go beyond a reasonable point */
		if world.CameraX > float32(def.XYMax) {
			world.CameraX = float32(def.XYMax)
		} else if world.CameraX < def.XYMin {
			world.CameraX = def.XYMin
		}
		if world.CameraY > float32(def.XYMax) {
			world.CameraY = float32(def.XYMax)
		} else if world.CameraY < def.XYMin {
			world.CameraY = def.XYMin
		}
	} else {
		gCameraDrag = false
	}
}

/* Detect and record right click state */
func getRightMouseClicks() {
	defer util.ReportPanic("getRightMouseClicks")
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		gRightMouseHeld = false
		gLastActionPosition = world.XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		gRightMouseHeld = true
	}
}

func getMiddleMouseClicks() {
	defer util.ReportPanic("getMiddleMouseClicks")
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = false
		gLastActionPosition = world.XY{}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = true
	}
}

func rotateWorldObjects() {
	defer util.ReportPanic("rotateWorldObjects")
	/* Rotate object */
	if !gClickCaptured && inpututil.IsKeyJustPressed(ebiten.KeyR) {

		pos := util.FloatXYToPosition(WorldMouseX, WorldMouseY)

		chunk := util.GetChunk(pos)
		/* Valid chunk? */
		if chunk == nil {
			return
		}

		b := util.GetObj(pos, chunk)
		/* Valid building? */
		if b == nil {
			if SelectedItemType == def.MaxItemType {
				return
			}
			for pos := 0; pos < ToolbarMax; pos++ {
				if ToolbarItems[pos].OType == nil {
					continue
				}
				item := ToolbarItems[pos].OType
				if item.TypeI != SelectedItemType {
					continue
				}
				if !item.Rotatable {
					continue
				}

				dir := item.Direction
				if gShiftPressed {
					dir = util.RotCCW(dir)
				} else {
					dir = util.RotCW(dir)
				}
				item.Direction = dir

				DrawToolbar(true, false, pos)
			}

			return
		}

		/* Valid object? */
		if b.Obj == nil {
			return
		}

		RotateListAdd(b, !gShiftPressed, pos)
	}
}
