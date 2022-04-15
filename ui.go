package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/util"
	"fmt"
	"math"
	"math/rand"
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
			glob.LastPinch = dist
		}
		glob.PinchPressed = true
		glob.ZoomMouse = (glob.ZoomMouse + ((dist - glob.LastPinch) / 75))
		glob.LastPinch = dist
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
				glob.LastTouchA, glob.LastTouchB = util.MidPoint(tx, ty, ta, tb)

			} else {
				glob.LastTouchX = tx
				glob.LastTouchY = ty
			}
		}
		glob.TouchPressed = true

		if glob.PinchPressed {
			nx, ny := util.MidPoint(tx, ty, ta, tb)
			glob.CameraX = glob.CameraX + (float64(glob.LastTouchA-nx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(glob.LastTouchB-ny) / glob.ZoomScale)
			glob.LastTouchA, glob.LastTouchB = util.MidPoint(tx, ty, ta, tb)
		} else {
			glob.CameraX = glob.CameraX + (float64(glob.LastTouchX-tx) / glob.ZoomScale)
			glob.CameraY = glob.CameraY + (float64(glob.LastTouchY-ty) / glob.ZoomScale)
			glob.LastTouchX = tx
			glob.LastTouchY = ty
		}
	} else {
		glob.TouchPressed = false
	}

	/* Mouse scroll zoom */
	//Scroll zoom
	_, fsy := ebiten.Wheel()

	//Workaround for wasm mouse scroll being insane
	if glob.DetOS == consts.Wasm {
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
		glob.ZoomMouse = 63.46
		glob.ZoomSetup = true
	}
	glob.ZoomScale = ((glob.ZoomMouse * glob.ZoomMouse * glob.ZoomMouse) / 1000)

	//If scroll wheel, lock to sharp ratios when zoomed in, otherwise dont
	if !glob.PinchPressed {
		if glob.ZoomScale >= consts.SpriteScale {
			lockto := float64(consts.SpriteScale) / 2.0
			glob.ZoomScale = math.Round(glob.ZoomScale/lockto) * lockto
		} else {
			glob.ZoomScale = math.Round(glob.ZoomScale)
		}
	}

	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	mx := float64(intx)
	my := float64(inty)
	glob.MousePosX = mx
	glob.MousePosY = my
	var captured bool = false

	//Mouse clicks
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		glob.MousePressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		glob.MousePressed = true
		glob.LastObjPos = glob.XYEmpty
		glob.LastActionType = consts.DragActionTypeNone

		//Toolbar
		uipix := float64(objects.ToolbarMax * int(consts.TBSize))
		if glob.MousePosX <= uipix+consts.ToolBarOffsetX {
			if glob.MousePosY <= consts.TBSize+consts.ToolBarOffsetY {
				ipos := int((glob.MousePosX - consts.ToolBarOffsetX) / consts.TBSize)
				temp := objects.ToolbarItems[ipos].Link
				item := temp[objects.ToolbarItems[ipos].Key]

				//Actions
				if item.UIAction != nil {
					item.UIAction()

					fmt.Println("UI Action:", item.Name)
				} else {
					objects.SelectedItemType = objects.ToolbarItems[ipos].Key
					fmt.Println("Selected:", item.Name)
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
			dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
			dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

			pos := util.FloatXYToPosition(dtx, dty)

			if pos != glob.LastObjPos {
				if time.Since(glob.LastActionTime) > glob.BuildActionDelay {

					chunk := util.GetChunk(pos)

					//Make chunk if needed
					if chunk == nil {
						cpos := util.PosToChunkPos(pos)
						fmt.Println("Made chunk:", cpos)

						chunk = &glob.MapChunk{}
						glob.WorldMap[cpos] = chunk
						chunk.MObj = make(map[glob.Position]*glob.MObj)
					}
					//Make obj if needed
					o := chunk.MObj[pos]
					bypass := false
					if o == nil {
						//Prevent flopping between delete and create when dragging
						if glob.LastActionType == consts.DragActionTypeBuild || glob.LastActionType == consts.DragActionTypeNone {
							size := objects.GameObjTypes[objects.SelectedItemType].Size
							if size.X > 1 || size.Y > 1 {
								for tx := 0; tx < size.X; tx++ {
									for ty := 0; ty < size.Y; ty++ {
										if chunk.MObj[glob.Position{X: pos.X + tx, Y: pos.Y + ty}] != nil {
											fmt.Println("ERROR: Occupied.")
											bypass = true
										}
									}
								}
							}
							if !bypass {
								o = &glob.MObj{}
								chunk.MObj[pos] = o
							}
						}
					}
					if !bypass && o != nil {
						//Change obj type
						if o.Type == consts.ObjTypeNone {

							o.Type = objects.SelectedItemType
							o.TypeP = objects.GameObjTypes[o.Type]
							o.OutputDir = consts.DIR_EAST
							glob.WorldMapDirty = true

							fmt.Println("Made obj:", pos, o.TypeP.Name)

							objects.LinkObj(pos, o)

							o.Valid = true
							/* Temporary for testing */
							if o.Contents[consts.DIR_INTERNAL] == nil {
								o.Contents[consts.DIR_INTERNAL] = &glob.MatData{}
							}
							o.Contents[consts.DIR_INTERNAL].Type = consts.MAT_COAL
							o.Contents[consts.DIR_INTERNAL].TypeP = objects.MatTypes[consts.MAT_COAL]
							/* Temporary for testing */

							if o.TypeP.ObjUpdate != nil {
								if o.TypeP.ProcSeconds > 0 {
									//Process on a specifc ticks
									objects.ToProcQue(o, objects.WorldTick+1+uint64(rand.Intn(int(o.TypeP.ProcSeconds))))
								} else {
									//Eternal
									objects.ToProcQue(o, 0)
								}
							}
							objects.ToTickQue(o)
							objects.ToTockQue(o)

							//Create tick and tock events

							//Action completed, save position and time
							glob.LastObjPos = pos
							glob.LastActionType = consts.DragActionTypeBuild
							//glob.LastActionTime = time.Now()

						} else {
							if time.Since(glob.LastActionTime) > glob.RemoveActionDelay {
								if glob.LastActionType == consts.DragActionTypeDelete || glob.LastActionType == consts.DragActionTypeNone {
									//Delete object
									fmt.Println("Object deleted:", pos, o.TypeP.Name)

									//Invalidate and delete
									chunk.MObj[pos].Valid = false
									delete(chunk.MObj, pos)
									glob.WorldMapDirty = true

									//Action completed, save position and time
									glob.LastObjPos = pos
									glob.LastActionType = consts.DragActionTypeDelete
									//glob.LastActionTime = time.Now()

									//Delete chunk if empty
									if len(chunk.MObj) <= 0 {
										cpos := util.PosToChunkPos(pos)
										fmt.Println("Chunk deleted:", cpos)
										delete(glob.WorldMap, cpos)
									}
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
		if !glob.SetupMouse {
			glob.LastMouseX = mx
			glob.LastMouseY = my
			glob.SetupMouse = true
		}

		glob.CameraX = glob.CameraX + (float64(glob.LastMouseX-mx) / glob.ZoomScale)
		glob.CameraY = glob.CameraY + (float64(glob.LastMouseY-my) / glob.ZoomScale)

		glob.LastMouseX = mx
		glob.LastMouseY = my

		//log.Println(cameraX, cameraY)
	} else {
		glob.SetupMouse = false
	}

	//Toggle arrows
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		if glob.ShowAltView {
			glob.ShowAltView = false
		} else {
			glob.ShowAltView = true
		}
	}

	//Rotate object
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		//Get mouse position on world
		dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
		dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

		pos := util.FloatXYToPosition(dtx, dty)

		chunk := util.GetChunk(pos)
		o := chunk.MObj[pos]

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
			fmt.Println("Rotated output:", pos, o.TypeP.Name, o.OutputDir)
			objects.LinkObj(pos, o)
			glob.WorldMapDirty = true
		}
	}

	return nil
}
