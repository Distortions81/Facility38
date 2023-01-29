package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/util"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	/* Touch vars */
	gPrevTouchX   int
	gPrevTouchY   int
	gPrevTouchA   int
	gPrevTouchB   int
	gPrevPinch    float64
	gTouchPressed bool
	gPinchPressed bool
	gTouchZoom    float64

	/* UI state */
	gMouseHeld      bool
	gRightMouseHeld bool
	gShiftPressed   bool
	gClickCaptured  bool

	/* Mouse vars */
	gMouseX     float64
	gMouseY     float64
	gPrevMouseX float64
	gPrevMouseY float64

	/* Last object we performed an action on */
	gLastActionPosition glob.XY
	gLastActionTime     time.Time
	gBuildActionDelay   time.Duration
	gRemoveActionDelay  time.Duration
	gLastActionType     int
)

const (
	cDragActionTypeNone   = 0
	cDragActionTypeBuild  = 1
	cDragActionTypeDelete = 2
)

/* Input handler */
func (g *Game) Update() error {

	var keys []ebiten.Key
	/* Game start screen */
	if !glob.PlayerReady &&
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.AppendPressedKeys(keys) != nil) {
		glob.PlayerReady = true
		glob.AllowUI = true
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
		return nil
	}
	gClickCaptured = false

	getMouseClicks()
	getRightMouseClicks()
	getShiftToggle()
	getMousePos()

	//touchScreenHandle()
	zoomHandle()

	createWorldObjects()
	moveCamera()
	toggleOverlays()
	rotateWorldObjects()

	return nil
}

func getShiftToggle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		gShiftPressed = true
	} else if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		gShiftPressed = false
	}
}

func handleToolbar(rotate bool) bool {
	uipix := float64(ToolbarMax * int(consts.ToolBarScale))
	if glob.MouseX <= uipix+consts.ToolBarOffsetX {
		if glob.MouseY <= consts.ToolBarScale+consts.ToolBarOffsetY {
			ipos := int((glob.MouseX - consts.ToolBarOffsetX) / consts.ToolBarScale)
			item := ToolbarItems[ipos].OType

			/* Actions */
			if item.ToolbarAction != nil {
				item.ToolbarAction()
			} else {
				if rotate && objects.GameObjTypes[SelectedItemType] != nil {
					dir := objects.GameObjTypes[SelectedItemType].Direction
					if gShiftPressed {
						dir = dir - 1
						if dir < consts.DIR_NORTH {
							dir = consts.DIR_WEST
						}
					} else {
						dir = dir + 1
						if dir > consts.DIR_WEST {
							dir = consts.DIR_NORTH
						}
					}
					objects.GameObjTypes[SelectedItemType].Direction = dir
					DrawToolbar()
				} else if SelectedItemType == ToolbarItems[ipos].OType.TypeI {
					SelectedItemType = 0
					DrawToolbar()
				} else {
					SelectedItemType = ToolbarItems[ipos].OType.TypeI
					DrawToolbar()
				}
			}
			gMouseHeld = false
			return true
		}
	}
	return false
}

func touchScreenHandle() {
	/* Touchscreen input */
	tids := ebiten.TouchIDs()

	tx := 0
	ty := 0
	ta := 0
	tb := 0

	/* Find touch events */
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

	/* Touch zoom-pinch */
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
	/* Touch pan */
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
			glob.CameraX = glob.CameraX + (float64(gPrevTouchA-nx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(gPrevTouchB-ny) / glob.ZoomScale)
			gPrevTouchA, gPrevTouchB = util.MidPoint(tx, ty, ta, tb)
			glob.CameraDirty = true
		} else {
			glob.CameraX = glob.CameraX + (float64(gPrevTouchX-tx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(gPrevTouchY-ty) / glob.ZoomScale)
			gPrevTouchX = tx
			gPrevTouchY = ty
			glob.CameraDirty = true
		}
	} else {
		gTouchPressed = false
	}
}

/* WASM wierdness kludge */
var lastScroll time.Time

func zoomHandle() {
	/* Mouse scroll zoom */
	_, fsy := ebiten.Wheel()

	if glob.FixWASM && fsy != 0 {
		if time.Since(lastScroll) < time.Millisecond*200 {
			return
		}
	}
	lastScroll = time.Now()

	if fsy > 0 || inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		glob.ZoomScale = glob.ZoomScale * 2
		glob.CameraDirty = true
	} else if fsy < 0 || inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		glob.ZoomScale = glob.ZoomScale / 2
		glob.CameraDirty = true
	}
	gTouchZoom = 0

	if glob.ZoomScale < 1 {
		glob.ZoomScale = 1
		glob.CameraDirty = true
	} else if glob.ZoomScale > 256 {
		glob.ZoomScale = 256
		glob.CameraDirty = true
	}

}

func getMousePos() {
	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	gMouseX = float64(intx)
	gMouseY = float64(inty)
	glob.MouseX = gMouseX
	glob.MouseY = gMouseY
	gClickCaptured = false

}

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

func createWorldObjects() {
	if gMouseHeld {

		/* UI area */
		if !gClickCaptured {
			/* Get mouse position on world */
			worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
			worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

			pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

			if pos != gLastActionPosition {
				if time.Since(gLastActionTime) > gBuildActionDelay {

					bypass := false
					chunk := util.GetChunk(&pos)
					o := util.GetObj(&pos, chunk)

					if o == nil {

						/* Prevent flopping between delete and create when dragging */
						if gLastActionType == cDragActionTypeBuild || gLastActionType == cDragActionTypeNone {

							/*
								size := objects.GameObjTypes[objects.SelectedItemType].Size
								if size.X > 1 || size.Y > 1 {
									var tx, ty int
									for tx = 0; tx < size.X; tx++ {
										for ty = 0; ty < size.Y; ty++ {
											if chunk.LargeWObject[glob.XY{X: pos.X + tx, Y: pos.Y + ty}] != nil {
												fmt.Println("ERROR: Occupied.")
												bypass = true
											}
										}
									}
								}
							*/

							if !bypass {
								go func(o *glob.WObject, pos glob.XY) {
									objects.ListLock.Lock()
									dir := objects.GameObjTypes[SelectedItemType].Direction
									objects.ObjectHitlistAdd(o, SelectedItemType, &pos, false, dir)
									DrawToolbar()
									objects.ListLock.Unlock()
								}(o, pos)

								gLastActionPosition = pos
								gLastActionType = cDragActionTypeBuild
							}
						}
					} else {
						if time.Since(gLastActionTime) > gRemoveActionDelay {
							if gLastActionType == cDragActionTypeDelete || gLastActionType == cDragActionTypeNone {

								if o != nil {
									go func(o *glob.WObject, pos glob.XY) {
										objects.ListLock.Lock()
										objects.ObjectHitlistAdd(o, o.TypeI, &pos, true, 0)
										objects.ListLock.Unlock()
									}(o, pos)
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
	}
}

func moveCamera() {

	speed := consts.WALKSPEED
	if gShiftPressed {
		speed = consts.RUNSPEED
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		glob.CameraY -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		glob.CameraX -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		glob.CameraY += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		glob.CameraX += speed
	}

	/* Mouse pan */
	if gRightMouseHeld {
		if !glob.InitMouse {
			gPrevMouseX = gMouseX
			gPrevMouseY = gMouseY
			glob.InitMouse = true
		}

		glob.CameraX = glob.CameraX + (float64(gPrevMouseX-gMouseX) / glob.ZoomScale)
		glob.CameraY = glob.CameraY + (float64(gPrevMouseY-gMouseY) / glob.ZoomScale)
		glob.CameraDirty = true

		/* Don't let camera go beyond a reasonable point */
		if glob.CameraX > float64(consts.XYMax) {
			glob.CameraX = float64(consts.XYMax)
		} else if glob.CameraX < consts.XYMin {
			glob.CameraX = consts.XYMin
		}
		if glob.CameraY > float64(consts.XYMax) {
			glob.CameraY = float64(consts.XYMax)
		} else if glob.CameraY < consts.XYMin {
			glob.CameraY = consts.XYMin
		}

		gPrevMouseX = gMouseX
		gPrevMouseY = gMouseY
	} else {
		glob.InitMouse = false
	}
}

func getRightMouseClicks() {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		gRightMouseHeld = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		gRightMouseHeld = true
	}
}

func toggleOverlays() {
	/* Toggle info overlay */
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		if glob.ShowInfoLayer {
			glob.ShowInfoLayer = false
		} else {
			glob.ShowInfoLayer = true
		}
	}
}

func rotateWorldObjects() {
	/* Rotate object */
	if !gClickCaptured && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		/* Get mouse position on world */
		worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
		worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

		chunk := util.GetChunk(&pos)
		o := chunk.WObject[pos]

		if o != nil {

			if gShiftPressed {
				o.Direction = util.RotCW(o.Direction)
			} else {
				o.Direction = util.RotCCW(o.Direction)
			}

			o.OutputObj = nil
			objects.LinkObj(pos, o)
		}
	}
}
