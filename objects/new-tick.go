package objects

import "GameTest/world"

type OffsetData struct {
	Offset uint64
	Ticks  []NewTickEvent
	Tocks  []NewTickEvent
}

type NewTickEvent struct {
	Obj    *world.ObjData
	Offset uint64
}

type TickInterval struct {
	Interval uint64
	Offsets  []OffsetData
}

var TickIntervals []TickInterval

/* TO DO: Add items to these and test */

func NewRunTocks() {
	for _, ti := range TickIntervals {
		for _, off := range ti.Offsets {
			if GameTick%(ti.Interval+off.Offset) == 0 {
				for _, tock := range off.Tocks {
					tock.Obj.Unique.TypeP.UpdateObj(tock.Obj)
				}
			}
		}
	}
}

func NewRunTicks() {
	for _, ti := range TickIntervals {
		for _, off := range ti.Offsets {
			if GameTick%(ti.Interval+off.Offset) == 0 {
				for _, tock := range off.Tocks {
					tickObj(tock.Obj)
				}
			}
		}
	}
}
