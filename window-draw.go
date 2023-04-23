package main

import (
	"Facility38/world"
	"image/color"
)

func testWindow(window *WindowData) {
	DrawText("Test", world.BootFont, world.ColorRed, color.Transparent,
		world.XYf32{X: float32(window.Size.X / 2), Y: float32(window.Size.Y / 2)}, 0, window.Cache, false, false, false)
}
