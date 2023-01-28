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
	PrevTouchX int     = 0
	PrevTouchY int     = 0
	PrevTouchA int     = 0
	PrevTouchB int     = 0
	PrevPinch  float64 = 0

	/* UI state */
	MousePressed      bool = false
	MouseRightPressed bool = false
	TouchPressed      bool = false
	PinchPressed      bool = false
	ShiftPressed      bool = false

	/* Mouse vars */
	PrevMouseX float64 = 0
	PrevMouseY float64 = 0
	ZoomMouse  float64 = 0.0

	/* Last object we performed an action on */
	LastActionPosition glob.XY
	LastActionTime     time.Time
	BuildActionDelay   time.Duration = 0
	RemoveActionDelay  time.Duration = 0
	LastActionType     int           = 0
)

const (
	DragActionTypeNone   = 0
	DragActionTypeBuild  = 1
	DragActionTypeDelete = 2
)

func handleToolbar(rotate bool) bool {
	uipix := float64(objects.ToolbarMax * int(consts.ToolBarScale))
	if glob.MouseX <= uipix+consts.ToolBarOffsetX {
		if glob.MouseY <= consts.ToolBarScale+consts.ToolBarOffsetY {
			ipos := int((glob.MouseX - consts.ToolBarOffsetX) / consts.ToolBarScale)
			item := objects.ToolbarItems[ipos].OType

			/* Actions */
			if item.ToolbarAction != nil {
				item.ToolbarAction()
			} else {
				if rotate && objects.GameObjTypes[objects.SelectedItemType] != nil {
					dir := objects.GameObjTypes[objects.SelectedItemType].Direction
					if ShiftPressed {
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
					objects.GameObjTypes[objects.SelectedItemType].Direction = dir
				} else if objects.SelectedItemType == objects.ToolbarItems[ipos].OType.TypeI {
					objects.SelectedItemType = 0
				} else {
					objects.SelectedItemType = objects.ToolbarItems[ipos].OType.TypeI
				}
			}
			MousePressed = false
			return true
		}
	}
	return false
}

/* Ebiten main loop */
func (g *Game) Update() error {

	if !glob.DrewMap &&
		(inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)) {
		glob.DrewMap = true
		glob.BootImage.Dispose()
	}

	if consts.NoInterface {
		return nil
	}

	if !glob.DrewMap {
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		ShiftPressed = true
	} else if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		ShiftPressed = false
	}

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
		if !PinchPressed {
			PrevPinch = dist
		}
		PinchPressed = true
		ZoomMouse = (ZoomMouse + ((dist - PrevPinch) / 75))
		PrevPinch = dist
	} else {
		if PinchPressed {
			TouchPressed = false
			foundTouch = false
		}
		PinchPressed = false
	}
	/* Touch pan */
	if foundTouch {
		if !TouchPressed {
			if PinchPressed {
				PrevTouchA, PrevTouchB = util.MidPoint(tx, ty, ta, tb)

			} else {
				PrevTouchX = tx
				PrevTouchY = ty
			}
		}
		TouchPressed = true

		if PinchPressed {
			nx, ny := util.MidPoint(tx, ty, ta, tb)
			glob.CameraX = glob.CameraX + (float64(PrevTouchA-nx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(PrevTouchB-ny) / glob.ZoomScale)
			PrevTouchA, PrevTouchB = util.MidPoint(tx, ty, ta, tb)
			glob.CameraDirty = true
		} else {
			glob.CameraX = glob.CameraX + (float64(PrevTouchX-tx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(PrevTouchY-ty) / glob.ZoomScale)
			PrevTouchX = tx
			PrevTouchY = ty
			glob.CameraDirty = true
		}
	} else {
		TouchPressed = false
	}

	/* Mouse scroll zoom */
	_, fsy := ebiten.Wheel()

	if fsy > 0 || inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		glob.ZoomScale = glob.ZoomScale * 2
		glob.CameraDirty = true
	} else if fsy < 0 || inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		glob.ZoomScale = glob.ZoomScale / 2
		glob.CameraDirty = true
	}
	ZoomMouse = 0

	if glob.ZoomScale < 1 {
		glob.ZoomScale = 1
		glob.CameraDirty = true
	} else if glob.ZoomScale > 256 {
		glob.ZoomScale = 256
		glob.CameraDirty = true
	}

	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	mx := float64(intx)
	my := float64(inty)
	glob.MouseX = mx
	glob.MouseY = my
	var captured bool = false

	/* Mouse clicks */
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		MousePressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		MousePressed = true
		LastActionPosition.X = 0
		LastActionPosition.Y = 0
		LastActionType = DragActionTypeNone

		captured = handleToolbar(false)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		captured = handleToolbar(true)
	}

	if MousePressed {

		/* UI area */
		if !captured {
			/* Get mouse position on world */
			worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
			worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

			pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

			if pos != LastActionPosition {
				if time.Since(LastActionTime) > BuildActionDelay {

					bypass := false
					chunk := util.GetChunk(&pos)
					o := util.GetObj(&pos, chunk)

					if o == nil {

						/* Prevent flopping between delete and create when dragging */
						if LastActionType == DragActionTypeBuild || LastActionType == DragActionTypeNone {

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
									dir := objects.GameObjTypes[objects.SelectedItemType].Direction
									objects.ObjectHitlistAdd(o, objects.SelectedItemType, &pos, false, dir)
									objects.ListLock.Unlock()
								}(o, pos)

								LastActionPosition = pos
								LastActionType = DragActionTypeBuild
							}
						}
					} else {
						if time.Since(LastActionTime) > RemoveActionDelay {
							if LastActionType == DragActionTypeDelete || LastActionType == DragActionTypeNone {

								if o != nil {
									go func(o *glob.WObject, pos glob.XY) {
										objects.ListLock.Lock()
										objects.ObjectHitlistAdd(o, o.TypeI, &pos, true, 0)
										objects.ListLock.Unlock()
									}(o, pos)
									//Action completed, save position and time
									LastActionPosition = pos
									LastActionType = DragActionTypeDelete
								}
							}
						}

					}
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		MouseRightPressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		MouseRightPressed = true
	}

	/* Mouse pan */
	if MouseRightPressed {
		if !glob.InitMouse {
			PrevMouseX = mx
			PrevMouseY = my
			glob.InitMouse = true
		}

		glob.CameraX = glob.CameraX + (float64(PrevMouseX-mx) / glob.ZoomScale)
		glob.CameraY = glob.CameraY + (float64(PrevMouseY-my) / glob.ZoomScale)
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

		PrevMouseX = mx
		PrevMouseY = my
	} else {
		glob.InitMouse = false
	}

	/* Toggle info overlay */
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		if glob.ShowInfoLayer {
			glob.ShowInfoLayer = false
		} else {
			glob.ShowInfoLayer = true
		}
	}

	/* Rotate object */
	if !captured && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		/* Get mouse position on world */
		worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
		worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

		chunk := util.GetChunk(&pos)
		o := chunk.WObject[pos]

		if o != nil {

			if ShiftPressed {
				o.Direction = util.RotCW(o.Direction)
			} else {
				o.Direction = util.RotCCW(o.Direction)
			}

			o.OutputObj = nil
			objects.LinkObj(pos, o)
		}
	}

	return nil
}
