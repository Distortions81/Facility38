package objects

import (
	"GameTest/cwlog"
	"GameTest/world"
)

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
		_, new := GetInterval(int(ot.TockInterval))
		if new {
			cwlog.DoLog(true, "Object: %v: Interval: %v", ot.Name, ot.TockInterval)
		}
	}
	cwlog.DoLog(true, "%v intervals added.", len(TickIntervals))
}

/* Return interval data, or create it if needed */
func GetInterval(interval int) (pos int, created bool) {
	foundInterval := false

	/* Eventually replace with precalc table */
	for ipos, inter := range TickIntervals {
		if inter.Interval == interval {
			foundInterval = true
			return ipos, false
		}
	}
	if !foundInterval {
		pos := len(TickIntervals)

		offsets := make([]OffsetData, interval-1)
		TickIntervals = append(TickIntervals, TickInterval{Interval: interval, Offsets: offsets})
		return pos, true
	}
	return -1, false
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
