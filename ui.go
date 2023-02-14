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
	/* Touch vars
	gPrevTouchX   int
	gPrevTouchY   int
	gPrevTouchA   int
	gPrevTouchB   int
	gPrevPinch    float32
	gTouchPressed bool
	gPinchPressed bool
	gTouchZoom    float32 */

	/* UI state */
	gMouseHeld      bool
	gRightMouseHeld bool
	gShiftPressed   bool
	gClickCaptured  bool

	/* Mouse vars */
	gMouseX     float32 = 1
	gMouseY     float32 = 1
	gPrevMouseX float32 = 1
	gPrevMouseY float32 = 1

	/* Last object we performed an action on */
	gLastActionPosition world.XY
	gLastActionTime     time.Time
	gLastActionType     int

	/* WASM wierdness kludge */
	lastScroll time.Time
)

const (
	cDragActionTypeNone   = 0
	cDragActionTypeBuild  = 1
	cDragActionTypeDelete = 2
)

func init() {
	lastScroll = time.Now()
}

/* Input interface handler */
func (g *Game) Update() error {

	g.ui.Update()

	var keys []ebiten.Key
	/* Game start screen */
	if !world.PlayerReady.Load() &&
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.AppendPressedKeys(keys) != nil) {
		world.PlayerReady.Store(true)
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
		return nil
	}
	gClickCaptured = false

	getMouseClicks()
	getRightMouseClicks()
	getShiftToggle()
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

/* Quit if alt-f4 or ESC are pressed */
func handleQuit() {
	if (inpututil.IsKeyJustPressed(ebiten.KeyF4) && ebiten.IsKeyPressed(ebiten.KeyAlt)) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		util.Chat("Game closing...")
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
	uipix := float32(ToolbarMax * int(gv.ToolBarScale))

	if world.MouseX <= uipix {
		if world.MouseY <= gv.ToolBarScale {

			ipos := int((world.MouseX) / gv.ToolBarScale)
			item := ToolbarItems[ipos].OType

			/* Actions */
			if item.ToolbarAction != nil && !rotate {
				item.ToolbarAction()
				//util.Chat(item.Name)
			} else {
				if rotate && item != nil {
					dir := item.Direction
					if gShiftPressed {
						dir = util.RotCCW(dir)
					} else {
						dir = util.RotCW(dir)
					}
					item.Direction = dir
					DrawToolbar()
					/* Deselect */
				} else if SelectedItemType == ToolbarItems[ipos].OType.TypeI-1 {
					SelectedItemType = 255
					DrawToolbar()
				} else {
					/* Select */
					SelectedItemType = ToolbarItems[ipos].OType.TypeI - 1
					DrawToolbar()
				}
			}
			gMouseHeld = false
			return true
		}
	}
	return false
}

/* Touchscreen input, incomplete
func touchScreenHandle() {
	tids := ebiten.TouchIDs()

	tx := 0
	ty := 0
	ta := 0
	tb := 0

	foundTouch := false
	foundPinch := false
	for _, tid := range tids {
		ttx, tty := ebiten.TouchPosition(tid)
		if ttx > 0 || tty > 0 {
			if foundTouch {
				ta = ttx
				tb = tty
				foundPinch = true
				break
			} else {
				tx = ttx
				ty = tty
				foundTouch = true
			}

		}
	}

	if foundPinch {
		dist := util.Distance((ta), (tb), (tx), (ty))
		if !gPinchPressed {
			gPrevPinch = dist
		}
		gPinchPressed = true
		gTouchZoom = (gTouchZoom + ((dist - gPrevPinch) / 75))
		gPrevPinch = dist
	} else {
		if gPinchPressed {
			gTouchPressed = false
			foundTouch = false
		}
		gPinchPressed = false
	}

	if foundTouch {
		if !gTouchPressed {
			if gPinchPressed {
				gPrevTouchA, gPrevTouchB = util.MidPoint(tx, ty, ta, tb)

			} else {
				gPrevTouchX = tx
				gPrevTouchY = ty
			}
		}
		gTouchPressed = true

		if gPinchPressed {
			nx, ny := util.MidPoint(tx, ty, ta, tb)
			world.CameraX = world.CameraX + (float32(gPrevTouchA-nx) / world.ZoomScale)
			world.CameraY = world.CameraY + (float32(gPrevTouchB-ny) / world.ZoomScale)
			gPrevTouchA, gPrevTouchB = util.MidPoint(tx, ty, ta, tb)
			world.VisDataDirty.Store(true)
		} else {
			world.CameraX = world.CameraX + (float32(gPrevTouchX-tx) / world.ZoomScale)
			world.CameraY = world.CameraY + (float32(gPrevTouchY-ty) / world.ZoomScale)
			gPrevTouchX = tx
			gPrevTouchY = ty
			world.VisDataDirty.Store(true)
		}
	} else {
		gTouchPressed = false
	}
} */

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
		gLastActionType = cDragActionTypeNone

		gClickCaptured = handleToolbar(false)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		gClickCaptured = handleToolbar(true)
	}
}

/* Look for clicks in window, create or destroy objects */
func createWorldObjects() {
	if gMouseHeld {

		/* UI area */
		if !gClickCaptured {
			/* Get mouse position on world */
			worldMouseX := (world.MouseX/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
			worldMouseY := (world.MouseY/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))

			pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

			if pos != gLastActionPosition {

				bypass := false
				chunk := util.GetChunk(pos)
				o := util.GetObj(pos, chunk)

				if o == nil {

					/* Prevent flopping between delete and create when dragging */
					if gLastActionType == cDragActionTypeBuild || gLastActionType == cDragActionTypeNone {

						/*
							size := objects.GameObjTypes[objects.SelectedItemType].Size
							if size.X > 1 || size.Y > 1 {
								var tx, ty int
								for tx = 0; tx < size.X; tx++ {
									for ty = 0; ty < size.Y; ty++ {
										if chunk.LargeWObject[world.XY{X: pos.X + tx, Y: pos.Y + ty}] != nil {
											cwlog.DoLog("ERROR: Occupied.")
											bypass = true
										}
									}
								}
							}
						*/

						if !bypass {
							if SelectedItemType == 255 {
								return
							}
							dir := objects.GameObjTypes[SelectedItemType].Direction
							oPos := util.CenterXY(pos)
							util.ChatDetailed(fmt.Sprintf("Created %v at (%v,%v)", objects.GameObjTypes[SelectedItemType].Name, oPos.X, oPos.Y), color.White, time.Second*3)

							if gv.WASMMode {
								objects.ObjQueueAdd(o, SelectedItemType, pos, false, dir)
							} else {
								go objects.ObjQueueAdd(o, SelectedItemType, pos, false, dir)
							}

							gLastActionPosition = pos
							gLastActionType = cDragActionTypeBuild
						}
					}
				} else {

					if gLastActionType == cDragActionTypeDelete || gLastActionType == cDragActionTypeNone {

						if o != nil {
							oPos := util.CenterXY(pos)
							util.ChatDetailed(fmt.Sprintf("Deleted %v at (%v,%v)", o.TypeP.Name, oPos.X, oPos.Y), color.White, time.Second*3)

							if gv.WASMMode {
								objects.ObjQueueAdd(o, o.TypeP.TypeI, pos, true, 0)
							} else {
								go objects.ObjQueueAdd(o, o.TypeP.TypeI, pos, true, 0)
							}
							//Action completed, save position and time
							gLastActionPosition = pos
							gLastActionType = cDragActionTypeDelete
						}
					}

				}

			}
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

	/* Mouse pan */
	if gRightMouseHeld {
		if !world.InitMouse {
			gPrevMouseX = gMouseX
			gPrevMouseY = gMouseY
			world.InitMouse = true
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
		world.InitMouse = false
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

/* Toggle overylays when ALT is pressed */
func toggleOverlays() {
	/* Toggle info overlay */
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		if world.ShowInfoLayer {
			world.ShowInfoLayer = false
			util.Chat("Info overlay is now off.")
		} else {
			world.ShowInfoLayer = true
			util.Chat("Info overlay is now on.")
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
		o := chunk.ObjMap[pos]

		if o != nil {

			objects.UnlinkObj(o)
			var newdir uint8
			if gShiftPressed {
				newdir = util.RotCCW(o.Dir)
				util.RotatePortsCCW(o)
				oPos := util.CenterXY(o.Pos)
				util.ChatDetailed(fmt.Sprintf("Rotated %v counter-clockwise at (%v,%v)", o.TypeP.Name, oPos.X, oPos.Y), color.White, time.Second*3)
			} else {
				newdir = util.RotCW(o.Dir)
				util.RotatePortsCW(o)
				oPos := util.CenterXY(o.Pos)
				util.ChatDetailed(fmt.Sprintf("Rotated %v clockwise at (%v,%v)", o.TypeP.Name, oPos.X, oPos.Y), color.White, time.Second*3)
			}
			o.Dir = newdir
			objects.LinkObj(o)
		}
	}
}
