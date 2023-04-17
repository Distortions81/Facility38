package objects

import (
	"GameTest/cwlog"
	"GameTest/world"
)

type OffsetData struct {
	Offset int
	Ticks  []*world.ObjData
	Tocks  []*world.ObjData
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
		_, new := GetIntervalPos(int(ot.TockInterval))
		if new {
			cwlog.DoLog(true, "Object: %v: Interval: %v", ot.Name, ot.TockInterval)
		}
	}
	cwlog.DoLog(true, "%v intervals added.", len(TickIntervals))
}

/* Return interval data, or create it if needed */
func GetIntervalPos(interval int) (pos int, created bool) {
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

		offsets := make([]OffsetData, interval+1)
		TickIntervals = append(TickIntervals, TickInterval{Interval: interval, Offsets: offsets})
		return pos, true
	}

	cwlog.DoLog(true, "Error!")
	return -1, false
}

func AddTock(obj *world.ObjData) {
	if obj.HasTock {
		return
	}
	i, _ := GetIntervalPos(int(obj.Unique.TypeP.TockInterval))

	if TickIntervals[i].LastOffset >= TickIntervals[i].Interval {
		TickIntervals[i].LastOffset = 0
	}

	TickIntervals[i].Offsets[TickIntervals[i].LastOffset].Tocks =
		append(TickIntervals[i].Offsets[TickIntervals[i].LastOffset].Tocks, obj)

	TickIntervals[i].LastOffset++
	world.TockCount++
}

func RemoveTock(obj *world.ObjData) {
	if !obj.HasTock {
		return
	}
	i, _ := GetIntervalPos(int(obj.Unique.TypeP.TockInterval))

	for offPos, off := range TickIntervals[i].Offsets {
		/* Check if this is the correct interval */
		if uint8(off.Offset) != obj.Unique.TypeP.TockInterval {
			continue
		}
		/* If it is, remove object */
		for itemPos, item := range off.Tocks {
			if item == obj {
				TickIntervals[i].Offsets[offPos].Tocks =
					append(
						TickIntervals[i].Offsets[offPos].Tocks[:itemPos],
						TickIntervals[i].Offsets[offPos].Tocks[itemPos+1:]...)

				world.TockCount--
				cwlog.DoLog(true, "Tock Removed: %v", obj.Unique.TypeP.Name)
				break
			}
		}
	}
}
func AddTick(obj *world.ObjData) {
	if obj.HasTick {
		return
	}
	i, _ := GetIntervalPos(int(obj.Unique.TypeP.TockInterval))
	if TickIntervals[i].LastOffset >= TickIntervals[i].Interval {
		TickIntervals[i].LastOffset = 0
	}

	TickIntervals[i].Offsets[TickIntervals[i].LastOffset].Ticks =
		append(TickIntervals[i].Offsets[TickIntervals[i].LastOffset].Ticks, obj)

	TickIntervals[i].LastOffset++
	world.TickCount++
}

func RemoveTick(obj *world.ObjData) {
	if !obj.HasTick {
		return
	}
	i, _ := GetIntervalPos(int(obj.Unique.TypeP.TockInterval))

	for offPos, off := range TickIntervals[i].Offsets {
		/* Check if this is the correct interval */
		if uint8(off.Offset) != obj.Unique.TypeP.TockInterval {
			continue
		}
		/* If it is, remove object */
		for itemPos, item := range off.Ticks {
			if item == obj {
				TickIntervals[i].Offsets[offPos].Ticks =
					append(
						TickIntervals[i].Offsets[offPos].Ticks[:itemPos],
						TickIntervals[i].Offsets[offPos].Ticks[itemPos+1:]...)

				world.TickCount--
				cwlog.DoLog(true, "Tick Removed: %v", obj.Unique.TypeP.Name)
				break
			}
		}
	}
}

func NewRunTocksST() {
	for _, ti := range TickIntervals {
		for _, off := range ti.Offsets {
			if ti.Interval == 0 || GameTick%uint64(ti.Interval+off.Offset) == 0 {
				for _, tock := range off.Tocks {
					tock.Unique.TypeP.UpdateObj(tock)
				}
			}
		}
	}
}

func NewRunTicksST() {
	for _, ti := range TickIntervals {
		for _, off := range ti.Offsets {
			if ti.Interval == 0 || GameTick%uint64(ti.Interval+off.Offset) == 0 {
				for _, tock := range off.Tocks {
					tickObj(tock)
				}
			}
		}
	}
}
