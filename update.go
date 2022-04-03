package main

import (
	"GameTest/GameWorld"
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//Ebiten main loop
func (g *Game) Update() error {

	if !glob.DrewMap {
		return nil
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
	} else if glob.ZoomMouse < 0.2 {
		glob.ZoomMouse = 0.2
	}
	if !glob.ZoomSetup {
		glob.ZoomMouse = 35
		glob.ZoomSetup = true
	}
	glob.ZoomScale = ((glob.ZoomMouse * glob.ZoomMouse * glob.ZoomMouse) / 4000)

	//If scroll wheel, lock to sharp ratios when zoomed in, otherwise dont
	/*if !glob.PinchPressed {
		if glob.ZoomScale >= 1 {
			glob.ZoomScale = math.Round(glob.ZoomScale)
		} else {
			glob.ZoomScale = math.Round(glob.ZoomScale*10) / 10
		}
	}*/

	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	mx := float64(intx)
	my := float64(inty)
	glob.MousePosX = mx
	glob.MousePosY = my

	//Mouse clicks
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		glob.MousePressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if !glob.MousePressed {
			glob.MousePressed = true

			captured := false

			//UI area
			//Toolbar
			if glob.MousePosX <= float64(glob.ObjTypeMax*int(glob.TBSize))+glob.ToolBarOffsetX {
				if glob.MousePosY <= glob.TBSize+glob.ToolBarOffsetY {
					itemType := int((glob.MousePosX-glob.ToolBarOffsetX)/glob.TBSize) + 1
					if glob.ObjTypes[itemType].GameObj {
						glob.SelectedItemType = itemType
					} else if glob.ObjTypes[itemType].Action != nil {
						glob.ObjTypes[itemType].Action()
					}
					captured = true
				}
			}

			if !captured {
				//Get mouse position on world
				dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
				dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))
				//Get position on game world
				gwx := (dtx / glob.DrawScale)
				gwy := (dty / glob.DrawScale)

				pos := util.FloatXYToPosition(gwx, gwy)

				chunk := util.GetChunk(pos)

				//Make chunk if needed
				if chunk == nil {
					cpos := util.PosToChunkPos(pos)
					fmt.Println("Made chunk:", cpos)

					chunk = &glob.MapChunk{}
					glob.WorldMap[cpos] = chunk
					chunk.MObj = make(map[glob.Position]*glob.MObj)
				}
				obj := chunk.MObj[pos]
				if obj == nil {
					fmt.Println("Made obj:", pos)
					obj = &glob.MObj{}
					chunk.MObj[pos] = obj
				}
				if obj.Type == glob.ObjTypeNone {
					obj.Type = glob.SelectedItemType
				} else {
					//Delete object
					fmt.Println("Object deleted:", pos)
					delete(chunk.MObj, pos)

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

	GameWorld.Update()
	return nil
}
