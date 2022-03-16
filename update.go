package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//Ebiten main loop
func (g *Game) Update() error {

	if glob.DrewMap == false {
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
	} else if glob.ZoomMouse < 6 {
		glob.ZoomMouse = 6
	}
	if !glob.ZoomSetup {
		glob.ZoomMouse = 35
		glob.ZoomSetup = true
	}
	glob.ZoomScale = ((glob.ZoomMouse * glob.ZoomMouse * glob.ZoomMouse) / 4000)

	//If scroll wheel, lock to sharp ratios when zoomed in, otherwise dont
	if !glob.PinchPressed {
		if glob.ZoomScale >= 1 {
			glob.ZoomScale = math.Round(glob.ZoomScale)
		} else {
			glob.ZoomScale = math.Round(glob.ZoomScale*10) / 10
		}
	}

	//Mouse clicks
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		glob.MousePressed = false
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		glob.MousePressed = true
	}

	/* Mouse position */
	intx, inty := ebiten.CursorPosition()
	mx := float64(intx)
	my := float64(inty)
	glob.MousePosX = mx
	glob.MousePosY = my

	/* Mouse pan */
	if glob.MousePressed {
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
	return nil
}
