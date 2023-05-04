package main

func linkMiner(obj *ObjData) {
	defer reportPanic("linkMiner")
	if obj.numOut == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	}
}

func linkBelt(obj *ObjData) {
	defer reportPanic("linkBelt")
	if obj.numOut == 0 || obj.numIn == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	}

	if obj.numIn == 1 && obj.numOut == 1 {
		var in, out uint8 = obj.inputs[0].Dir, obj.outputs[0].Dir

		obj.isCorner = true

		var DrawDir uint8 = DIR_NORTH
		if in == DIR_SOUTH && out == DIR_EAST ||
			out == DIR_SOUTH && in == DIR_EAST {
			DrawDir = 0
		} else if in == DIR_WEST && out == DIR_SOUTH ||
			out == DIR_WEST && in == DIR_SOUTH {
			DrawDir = 1
		} else if in == DIR_WEST && out == DIR_NORTH ||
			out == DIR_WEST && in == DIR_NORTH {
			DrawDir = 2
		} else if in == DIR_NORTH && out == DIR_EAST ||
			out == DIR_NORTH && in == DIR_EAST {
			DrawDir = 3
		} else {
			obj.isCorner = false
		}
		obj.cornerDir = DrawDir
	} else {
		obj.isCorner = false
	}
}

func linkBeltOver(obj *ObjData) {
	defer reportPanic("linkBeltOver")
	if obj.numOut == 0 || obj.numIn == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	}

	/* Alias inputs */
	for i, input := range obj.inputs {
		if input.Dir == reverseDirection(obj.Dir) {
			obj.beltOver.overIn = obj.inputs[i]
		} else {
			obj.beltOver.underIn = obj.inputs[i]
		}
	}

	/* Alias outputs */
	for o, output := range obj.outputs {
		if output.Dir == obj.Dir {
			obj.beltOver.overOut = obj.outputs[o]
		} else {
			obj.beltOver.underOut = obj.outputs[o]
		}
	}
}

func linkFuelHopper(obj *ObjData) {
	defer reportPanic("linkFuelHopper")
	if obj.numFOut == 0 || obj.numIn == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	}
}

func linkSplitter(obj *ObjData) {
	defer reportPanic("linkSplitter")
	if obj.numOut == 0 || obj.numIn == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	}
}

func linkUnloader(obj *ObjData) {
	defer reportPanic("linkUnloader")
	if obj.numOut != 0 || obj.numIn != 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	}
}

func linkBox(obj *ObjData) {
	defer reportPanic("linkBox")
	if obj.numIn == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
	}
}

func linkMachine(obj *ObjData) {
	defer reportPanic("linkMachine")
	if obj.numOut == 0 {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
	} else {
		eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
	}
}
