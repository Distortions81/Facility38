package objects

import "GameTest/world"

type OffsetData struct {
	Offset int
	Ticks  []NewTickEvent
	Tocks  []NewTickEvent
}

type NewTickEvent struct {
	Obj    *world.ObjData
	Offset uint8
}

type TickInterval struct {
	Interval   int
	LastOffset int
	Offsets    []OffsetData
}

var TickIntervals []TickInterval

/* Init at boot */
func init() {
	for _, ot := range WorldObjs {
		GetInterval(int(ot.TockInterval))
	}
}

/* Return interval data, or create it if needed */
func GetInterval(interval int) int {
	foundInterval := false

	/* Eventually replace with precalc table */
	for ipos, inter := range TickIntervals {
		if inter.Interval == interval {
			foundInterval = true
			return ipos
		}
	}
	if !foundInterval {
		pos := len(TickIntervals)
		TickIntervals = append(TickIntervals, TickInterval{Interval: interval})
		return pos
	}
	return -1
}

func AddTock(obj *world.ObjData) {
	//interval := GetInterval(int(obj.Unique.TypeP.TockInterval))

}

func NewRunTocksST() {
	for _, ti := range TickIntervals {
		for _, off := range ti.Offsets {
			if GameTick%uint64(ti.Interval+off.Offset) == 0 {
				for _, tock := range off.Tocks {
					tock.Obj.Unique.TypeP.UpdateObj(tock.Obj)
				}
			}
		}
	}
}

func NewRunTicksST() {
	for _, ti := range TickIntervals {
		for _, off := range ti.Offsets {
			if GameTick%uint64(ti.Interval+off.Offset) == 0 {
				for _, tock := range off.Tocks {
					tickObj(tock.Obj)
				}
			}
		}
	}
}
