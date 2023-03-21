package objects

import (
	"GameTest/gv"
	"GameTest/world"
)

func linkMiner(obj *world.ObjData) {
	if obj.NumOut == 0 {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
	}
}

func linkBelt(obj *world.ObjData) {
	if obj.NumOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
	}
}

func linkFuelHopper(obj *world.ObjData) {
	if obj.NumFOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
	}
}

func linkSplitter(obj *world.ObjData) {
	if obj.NumOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
	}
}

func linkBox(obj *world.ObjData) {
	if obj.NumIn == 0 {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
	} else {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
	}
}

func linkSmelter(obj *world.ObjData) {
	if obj.NumOut == 0 {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
	}
}
