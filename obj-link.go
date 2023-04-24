package main

import (
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
)

func linkMiner(obj *world.ObjData) {
	if obj.NumOut == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	}
}

func linkBelt(obj *world.ObjData) {
	if obj.NumOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	}

	if obj.NumIn == 1 && obj.NumOut == 1 {
		var in, out uint8 = obj.Inputs[0].Dir, obj.Outputs[0].Dir

		obj.IsCorner = true

		var DrawDir uint8 = def.DIR_NORTH
		if in == def.DIR_SOUTH && out == def.DIR_EAST ||
			out == def.DIR_SOUTH && in == def.DIR_EAST {
			DrawDir = 0
		} else if in == def.DIR_WEST && out == def.DIR_SOUTH ||
			out == def.DIR_WEST && in == def.DIR_SOUTH {
			DrawDir = 1
		} else if in == def.DIR_WEST && out == def.DIR_NORTH ||
			out == def.DIR_WEST && in == def.DIR_NORTH {
			DrawDir = 2
		} else if in == def.DIR_NORTH && out == def.DIR_EAST ||
			out == def.DIR_NORTH && in == def.DIR_EAST {
			DrawDir = 3
		} else {
			obj.IsCorner = false
		}
		obj.CornerDir = DrawDir
	}
}

func linkBeltOver(obj *world.ObjData) {
	if obj.NumOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	}

	/* Alias inputs */
	for i, input := range obj.Inputs {
		if input.Dir == util.ReverseDirection(obj.Dir) {
			obj.BeltOver.OverIn = obj.Inputs[i]
		} else {
			obj.BeltOver.UnderIn = obj.Inputs[i]
		}
	}

	/* Alias outputs */
	for o, output := range obj.Outputs {
		if output.Dir == obj.Dir {
			obj.BeltOver.OverOut = obj.Outputs[o]
		} else {
			obj.BeltOver.UnderOut = obj.Outputs[o]
		}
	}
}

func linkFuelHopper(obj *world.ObjData) {
	if obj.NumFOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	}
}

func linkSplitter(obj *world.ObjData) {
	if obj.NumOut == 0 || obj.NumIn == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	}
}

func linkUnloader(obj *world.ObjData) {
	if obj.NumOut != 0 || obj.NumIn != 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	}
}

func linkBox(obj *world.ObjData) {
	if obj.NumIn == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
	}
}

func linkMachine(obj *world.ObjData) {
	if obj.NumOut == 0 {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, true)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, true)
	} else {
		EventQueueAdd(obj, def.QUEUE_TYPE_TOCK, false)
		EventQueueAdd(obj, def.QUEUE_TYPE_TICK, false)
	}
}
