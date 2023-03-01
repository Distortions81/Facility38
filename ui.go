package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"image/color"
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

	/* Mouse vars */
	gMouseX     float32 = 1
	gMouseY     float32 = 1
	gPrevMouseX float32 = 1
	gPrevMouseY float32 = 1

	/* Last object we performed an action on */
	gLastActionPosition world.XY
	gLastKey            ebiten.Key

	/* WASM weirdness kludge */
	lastScroll time.Time
)

func init() {
	lastScroll = time.Now()
}

/* Input interface handler */
func (g *Game) Update() error {

	var keys []ebiten.Key
	/* Game start screen */
	if !world.PlayerReady.Load() &&
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
	getMousePos()

	handleQuit()

	//touchScreenHandle()
	zoomHandle()

	createWorldObjects()
	moveCamera()
	toggleOverlays()
	rotateWorldObjects()

	return nil
}

func getToolbarKeypress() {

	for _, item := range objects.UIObjsTypes {
		if item.QKey != 0 {
			if inpututil.IsKeyJustPressed(item.QKey) {
				if item.QKey != gLastKey {
					gLastKey = item.QKey

					item.ToolbarAction()
				}
			}
		}
	}
}

/* Quit if alt-f4 or ESC are pressed */
func handleQuit() {
	if (inpututil.IsKeyJustPressed(ebiten.KeyF4) && ebiten.IsKeyPressed(ebiten.KeyAlt)) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		util.ChatDetailed("Game closing...", world.ColorRed, time.Second*10)
		time.Sleep(time.Second * 5)
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

	if world.MouseX <= uipix {
		if world.MouseY <= gv.ToolBarScale {

			ipos := int((world.MouseX) / float32(gv.ToolBarScale+gv.ToolBarSpacing))
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
				} else if SelectedItemType == ToolbarItems[ipos].OType.TypeI-1 {
					SelectedItemType = 255
					DrawToolbar(true, false, ipos)
				} else {
					/* Select */
					SelectedItemType = ToolbarItems[ipos].OType.TypeI - 1
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

	if fsy > 0 || inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		world.ZoomScale = world.ZoomScale * 2
		world.VisDataDirty.Store(true)
	} else if fsy < 0 || inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		world.ZoomScale = world.ZoomScale / 2
		world.VisDataDirty.Store(true)
	}

	if world.ZoomScale < 1 {
		world.ZoomScale = 1
		world.VisDataDirty.Store(true)
	} else if world.ZoomScale > 256 {
		world.ZoomScale = 256
		world.VisDataDirty.Store(true)
	}

}

/* Get mos position and record it to world.MouseX/Y */
func getMousePos() {
	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	gMouseX = float32(intx)
	gMouseY = float32(inty)
	world.MouseX = gMouseX
	world.MouseY = gMouseY
	gClickCaptured = false

}

/* Record mouse clicks, send clicks to toolbar */
func getMouseClicks() {
	/* Mouse clicks */
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		gMouseHeld = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		gMouseHeld = true
		gLastActionPosition.X = 0
		gLastActionPosition.Y = 0

		gClickCaptured = handleToolbar(false)
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

	/* Get mouse position on world */
	worldMouseX := (world.MouseX/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
	worldMouseY := (world.MouseY/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))

	pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

	/* Is this a new position? */
	if pos == gLastActionPosition {
		return
	}

	chunk := util.GetChunk(pos)
	b := util.GetObj(pos, chunk)

	if b == nil && gMouseHeld {
		if SelectedItemType == 255 {
			return
		}
		dir := objects.GameObjTypes[SelectedItemType].Direction

		if gv.WASMMode {
			objects.ObjQueueAdd(nil, SelectedItemType, pos, false, dir)
		} else {
			go objects.ObjQueueAdd(nil, SelectedItemType, pos, false, dir)
		}
	} else if b != nil && b.Obj != nil && gRightMouseHeld {
		if gv.WASMMode {
			objects.ObjQueueAdd(b.Obj, SelectedItemType, pos, true, 0)
		} else {
			go objects.ObjQueueAdd(b.Obj, SelectedItemType, pos, true, 0)
		}
	}

	gLastActionPosition = pos
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
		if !world.MouseDrag {
			gPrevMouseX = gMouseX
			gPrevMouseY = gMouseY
			world.MouseDrag = true
		}

		world.CameraX = world.CameraX + (float32(gPrevMouseX-gMouseX) / world.ZoomScale)
		world.CameraY = world.CameraY + (float32(gPrevMouseY-gMouseY) / world.ZoomScale)
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

		gPrevMouseX = gMouseX
		gPrevMouseY = gMouseY
	} else {
		world.MouseDrag = false
	}
}

/* Detect and record right click state */
func getRightMouseClicks() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		gRightMouseHeld = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		gRightMouseHeld = true
	}
}

func getMiddleMouseClicks() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		gMiddleMouseHeld = true
	}
}

/* Toggle overylays when ALT is pressed */
func toggleOverlays() {
	/* Toggle info overlay */
	if inpututil.IsKeyJustPressed(objects.UIObjsTypes[gv.ToolbarLayer].QKey) {
		if world.ShowInfoLayer {
			world.ShowInfoLayer = false
			util.ChatDetailed("Info overlay is now off.", world.ColorOrange, time.Second*5)
		} else {
			world.ShowInfoLayer = true
			util.ChatDetailed("Info overlay is now on.", world.ColorOrange, time.Second*5)
		}
	}
}

/* Rotate objects when R or SHIFT-R are pressed */
func rotateWorldObjects() {
	/* Rotate object */
	if !gClickCaptured && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		/* Get mouse position on world */
		worldMouseX := (world.MouseX/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth/2.0) / world.ZoomScale)))
		worldMouseY := (world.MouseY/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight/2.0))/world.ZoomScale))

		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

		chunk := util.GetChunk(pos)
		if chunk == nil {
			return
		}
		o := util.GetObj(pos, chunk)

		if o != nil {

			objects.UnlinkObj(o.Obj)
			var newdir uint8
			if gShiftPressed {
				newdir = util.RotCCW(o.Obj.Dir)
				oPos := util.CenterXY(o.Obj.Pos)
				for p := range o.Obj.Ports {
					o.Obj.Ports[p].Dir = util.RotCCW(o.Obj.Dir)
				}
				util.ChatDetailed(fmt.Sprintf("Rotated %v counter-clockwise at (%v,%v)", o.Obj.TypeP.Name, oPos.X, oPos.Y), color.White, time.Second*5)
			} else {
				newdir = util.RotCW(o.Obj.Dir)
				oPos := util.CenterXY(o.Obj.Pos)
				for p := range o.Obj.Ports {
					o.Obj.Ports[p].Dir = util.RotCW(o.Obj.Dir)
				}
				util.ChatDetailed(fmt.Sprintf("Rotated %v clockwise at (%v,%v)", o.Obj.TypeP.Name, oPos.X, oPos.Y), color.White, time.Second*5)
			}
			o.Obj.Dir = newdir
			objects.LinkObj(o)
		}
	}
}
