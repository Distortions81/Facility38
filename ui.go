package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/util"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//Ebiten main loop
func (g *Game) Update() error {

	if !glob.DrewMap {
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		glob.ShiftPressed = true
	} else if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		glob.ShiftPressed = false
	}

	//Touchscreen input
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
		if !glob.PinchPressed {
			glob.PrevPinch = dist
		}
		glob.PinchPressed = true
		glob.ZoomMouse = (glob.ZoomMouse + ((dist - glob.PrevPinch) / 75))
		glob.PrevPinch = dist
	} else {
		if glob.PinchPressed {
			glob.TouchPressed = false
			foundTouch = false
		}
		glob.PinchPressed = false
	}
	/* Touch pan */
	if foundTouch {
		if !glob.TouchPressed {
			if glob.PinchPressed {
				glob.PrevTouchA, glob.PrevTouchB = util.MidPoint(tx, ty, ta, tb)

			} else {
				glob.PrevTouchX = tx
				glob.PrevTouchY = ty
			}
		}
		glob.TouchPressed = true

		if glob.PinchPressed {
			nx, ny := util.MidPoint(tx, ty, ta, tb)
			glob.CameraX = glob.CameraX + (float64(glob.PrevTouchA-nx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(glob.PrevTouchB-ny) / glob.ZoomScale)
			glob.PrevTouchA, glob.PrevTouchB = util.MidPoint(tx, ty, ta, tb)
		} else {
			glob.CameraX = glob.CameraX + (float64(glob.PrevTouchX-tx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(glob.PrevTouchY-ty) / glob.ZoomScale)
			glob.PrevTouchX = tx
			glob.PrevTouchY = ty
		}
	} else {
		glob.TouchPressed = false
	}

	/* Mouse scroll zoom */
	//Scroll zoom
	_, fsy := ebiten.Wheel()

	//Workaround for wasm mouse scroll being insane
	if glob.DetectedOS == consts.Wasm {
		glob.ZoomMouse = (glob.ZoomMouse + (fsy / 50))
	} else {
		glob.ZoomMouse = (glob.ZoomMouse + (fsy))
	}

	/* Zoom limits */
	if glob.ZoomMouse > 100 {
		glob.ZoomMouse = 100
	} else if glob.ZoomMouse < 10 {
		glob.ZoomMouse = 10
	}
	if !glob.ZoomSetup {
		glob.ZoomMouse = 47
		glob.ZoomSetup = true
	}
	glob.ZoomScale = ((glob.ZoomMouse * glob.ZoomMouse * glob.ZoomMouse) / 1000)

	//If scroll wheel, lock to sharp ratios when zoomed in, otherwise dont
	if !glob.PinchPressed {
		if glob.ZoomScale >= consts.SpriteScale {
			lockto := float64(consts.SpriteScale) / 2.0
			glob.ZoomScale = math.Floor(glob.ZoomScale/lockto) * lockto
		} else {
			glob.ZoomScale = math.Floor(glob.ZoomScale)
		}
	}

	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	mx := float64(intx)
	my := float64(inty)
	glob.MouseX = mx
	glob.MouseY = my
	var captured bool = false

	//Mouse clicks
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		glob.MousePressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		glob.MousePressed = true
		glob.LastActionPosition = glob.XYEmpty
		glob.LastActionType = consts.DragActionTypeNone

		//Toolbar
		uipix := float64(objects.ToolbarMax * int(consts.ToolBarScale))
		if glob.MouseX <= uipix+consts.ToolBarOffsetX {
			if glob.MouseY <= consts.ToolBarScale+consts.ToolBarOffsetY {
				ipos := int((glob.MouseX - consts.ToolBarOffsetX) / consts.ToolBarScale)
				item := objects.ToolbarItems[ipos].OType

				//Actions
				if item.ToolbarAction != nil {
					item.ToolbarAction()

					//fmt.Println("UI Action:", item.Name)
				} else {
					if objects.SelectedItemType == objects.ToolbarItems[ipos].OType.TypeI {
						objects.SelectedItemType = 0
						//fmt.Println("Deselected")
					} else {
						objects.SelectedItemType = objects.ToolbarItems[ipos].OType.TypeI
						//fmt.Println("Selected:", item.Name)
					}
				}
				captured = true
				glob.MousePressed = false
			}
		}
	}

	if glob.MousePressed {

		//UI area
		if !captured {
			//Get mouse position on world
			worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
			worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

			pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

			if pos != glob.LastActionPosition {
				if time.Since(glob.LastActionTime) > glob.BuildActionDelay {

					bypass := false
					chunk := util.GetChunk(&pos)
					o := util.GetObj(&pos, chunk)

					if o == nil {

						//Prevent flopping between delete and create when dragging
						if glob.LastActionType == consts.DragActionTypeBuild || glob.LastActionType == consts.DragActionTypeNone {

							size := objects.GameObjTypes[objects.SelectedItemType].Size
							if size.X > 1 || size.Y > 1 {
								var tx, ty int
								for tx = 0; tx < size.X; tx++ {
									for ty = 0; ty < size.Y; ty++ {
										if chunk.CObj[glob.Position{X: pos.X + tx, Y: pos.Y + ty}] != nil {
											//fmt.Println("ERROR: Occupied.")
											bypass = true
										}
									}
								}
							}

							if !bypass {
								objects.ObjectHitlistAdd(o, objects.SelectedItemType, &pos, false)

								glob.LastActionPosition = pos
								glob.LastActionType = consts.DragActionTypeBuild
							}
						}
					} else {
						if time.Since(glob.LastActionTime) > glob.RemoveActionDelay {
							if glob.LastActionType == consts.DragActionTypeDelete || glob.LastActionType == consts.DragActionTypeNone {

								if o != nil {
									objects.ObjectHitlistAdd(o, o.TypeI, &pos, true)
									//Action completed, save position and time
									glob.LastActionPosition = pos
									glob.LastActionType = consts.DragActionTypeDelete
									//glob.LastActionTime = time.Now()
								}
							}
						}

					}
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		glob.MouseRightPressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		glob.MouseRightPressed = true
	}

	/* Mouse pan */
	if glob.MouseRightPressed {
		if !glob.InitMouse {
			glob.PrevMouseX = mx
			glob.PrevMouseY = my
			glob.InitMouse = true
		}

		glob.CameraX = glob.CameraX + (float64(glob.PrevMouseX-mx) / glob.ZoomScale)
		glob.CameraY = glob.CameraY + (float64(glob.PrevMouseY-my) / glob.ZoomScale)

		//Max of 0 to 4,294,967,295
		if glob.CameraX > float64(consts.MaxUint) {
			glob.CameraX = float64(consts.MaxUint)
		} else if glob.CameraX < 0 {
			glob.CameraX = 0
		}
		if glob.CameraY > float64(consts.MaxUint) {
			glob.CameraY = float64(consts.MaxUint)
		} else if glob.CameraY < 0 {
			glob.CameraY = 0
		}

		glob.PrevMouseX = mx
		glob.PrevMouseY = my

		//log.Println(cameraX, cameraY)
	} else {
		glob.InitMouse = false
	}

	//Toggle arrows
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		if glob.ShowInfoLayer {
			glob.ShowInfoLayer = false
		} else {
			glob.ShowInfoLayer = true
		}
	}

	//Rotate object
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		//Get mouse position on world
		worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
		worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)

		chunk := util.GetChunk(&pos)
		o := chunk.WObject[pos]

		if o != nil {
			if glob.ShiftPressed {
				o.OutputDir = o.OutputDir - 1
				if o.OutputDir < consts.DIR_NORTH {
					o.OutputDir = consts.DIR_WEST
				}
			} else {
				o.OutputDir = o.OutputDir + 1
				if o.OutputDir > consts.DIR_WEST {
					o.OutputDir = consts.DIR_NORTH
				}
			}
			//fmt.Println("Rotated output:", pos, o.TypeP.Name, util.DirToName(o.OutputDir))
			objects.LinkObj(pos, o)
		}
	}

	return nil
}
