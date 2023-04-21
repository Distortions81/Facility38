package ui

import "Facility38/world"

type WindowData struct {
	Title string

	Movable    bool
	Autosized  bool
	Opaque     bool
	Scrollable bool
	Centered   bool

	WindowButtons WindowButtonData

	Size     world.XY
	Position world.XY
}

type WindowButtonData struct {
	Minimize bool
	CloseX   bool

	Cancel bool
	Okay   bool
	Save   bool
}

func DrawWindow(window *WindowData) {

}
