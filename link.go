package main

import (
	"fmt"
	"sync"
)

var linkLock sync.Mutex

/* Link to output in (dir) */
func linkObj(from XY, b *buildingData) {
	defer reportPanic("linkObj")
	linkLock.Lock()
	defer linkLock.Unlock()

	//ObjCD(b, fmt.Sprintf("Facing: %v", DirToName(b.Obj.Dir)))
	b.obj.LastInput = 0
	b.obj.LastOutput = 0

	/* Attempt to link ports */
	for p, port := range b.obj.Ports {

		/* Make sure port is unoccupied */
		if port.obj != nil {
			//ObjCD(b, fmt.Sprintf("Port Occupied: %v", DirToName(port.Dir)))
			continue
		}

		var neighb *buildingData
		if port.Dir == DIR_ANY {
			var testPort uint8
			for testPort = DIR_NORTH; testPort <= DIR_WEST; testPort++ {
				//DoLog(true, "Looking in all directions: "+DirToName(testPort))

				neighb = GetNeighborObj(from, testPort)
				if neighb != nil && neighb.obj != nil {
					if neighb.obj.Pos != b.obj.Pos {
						//DoLog(true, "found")
						break
					}
				}
			}
			if neighb == nil {
				continue
			}
		} else {
			neighb = GetNeighborObj(from, port.Dir)
		}

		/* We found one*/
		if neighb == nil {
			//ObjCD(b, fmt.Sprintf("No neighbor: %v", DirToName(port.Dir)))
			continue
		}

		if neighb.obj.Pos == b.obj.Pos {
			//ObjCD(b, fmt.Sprintf("Ignoring link to self: %v", DirToName(port.Dir)))
			continue
		}

		//DoLog(true, DirToName(port.Dir))

		for n, np := range neighb.obj.Ports {

			/* Neighbor port is available */
			if np.obj != nil {
				//ObjCD(b, fmt.Sprintf("Port occupied: %v", DirToName(port.Dir)))
				continue
			}

			/* Port is in correct direction */
			if np.Dir == ReverseDirection(port.Dir) ||
				np.Dir == DIR_ANY || port.Dir == DIR_ANY {

				/* Port is of correct type */
				if port.Type != ReverseType(np.Type) {
					//ObjCD(b, fmt.Sprintf("Port incorrect type: %v", DirToName(port.Dir)))
					continue
				}

				/* Normal objects can only link to loaders */
				if (b.obj.Unique.typeP.category == ObjCatGeneric &&
					neighb.obj.Unique.typeP.category != ObjCatLoader) ||
					(neighb.obj.Unique.typeP.category == ObjCatGeneric &&
						b.obj.Unique.typeP.category != ObjCatLoader) {
					continue
				}

				/* Add link to objects */
				neighb.obj.Ports[n].obj = b.obj
				b.obj.Ports[p].obj = neighb.obj

				/* Add direct port links */
				neighb.obj.Ports[n].link = &b.obj.Ports[p]
				b.obj.Ports[p].link = &neighb.obj.Ports[n]

				if Debug {
					oName := "none"
					if b.obj != nil {
						oName = fmt.Sprintf("%v: %v", neighb.obj.Unique.typeP.name, PosToString(neighb.pos))
					}
					ObjCD(b, fmt.Sprintf("Linked: Port-%v: ( %v %v ) to %v", p, DirToName(port.Dir), DirToArrow(port.Dir), oName))
				}

				portAlias(b.obj, p, port.Type)
				portAlias(neighb.obj, n, np.Type)

				/* Run custom link code */
				if neighb.obj.Unique.typeP.linkObj != nil {
					neighb.obj.Unique.typeP.linkObj(neighb.obj)
				} else {
					autoEvents(neighb.obj)
				}
			}
		}
		/* Run custom link code */
		if b.obj.Unique.typeP.linkObj != nil {
			b.obj.Unique.typeP.linkObj(b.obj)
		} else {
			autoEvents(b.obj)
		}

	}

}

/* Add/Remove tick/tock events as needed */
func autoEvents(obj *ObjData) {
	defer reportPanic("AutoEvents")

	/* Add to tock/tick lists */
	var foundOutputs, foundInputs, foundFOut, foundFIn bool
	if obj.numOut > 0 {
		foundOutputs = true
	}
	if obj.numIn > 0 {
		foundInputs = true
	}
	if obj.numFIn > 0 {
		foundFIn = true
	}
	if obj.numFOut > 0 {
		foundFOut = true
	}

	/* If we have inputs and outputs object needs, add to tock list */
	if obj.Unique.typeP.updateObj != nil {

		if obj.Unique.typeP.hasInputs && foundInputs {
			EventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}
		if obj.Unique.typeP.hasOutputs && foundOutputs {
			EventQueueAdd(obj, QUEUE_TYPE_TICK, false)
			EventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}

		if obj.Unique.typeP.hasFIn && foundFIn {
			EventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}
		if obj.Unique.typeP.hasFOut && foundFOut {
			EventQueueAdd(obj, QUEUE_TYPE_TICK, false)
			EventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}
	}
}

/* unlinkObj an object */
func unlinkObj(obj *ObjData) {
	defer reportPanic("UnlinkObj")
	linkLock.Lock()
	defer linkLock.Unlock()

	/* Reset last input var */
	obj.LastInput = 0
	obj.LastOutput = 0

	for p, port := range obj.Ports {
		/* No obj, skip */
		if port.obj == nil {
			continue
		}

		/* Delete ourselves from others */
		for pb, portb := range port.obj.Ports {
			if portb.obj == obj {
				pObj := port.obj

				/* Reset last port to avoid hitting invalid one */
				if port.Type == PORT_IN {
					obj.LastInput = 0
				} else {
					port.obj.LastInput = 0
				}

				/* Clean up port aliases */
				pObj.Ports[pb].link = nil
				pObj.Ports[pb].obj = nil

				portPop(pObj)
				if pObj.Unique.typeP.linkObj != nil {
					pObj.Unique.typeP.linkObj(pObj)
				} else {
					autoEvents(pObj)
				}
			}
		}
		portPop(port.obj)

		/* Break links */
		obj.Ports[p].link = nil
		obj.Ports[p].obj = nil
	}

	/* Murder all the port aliases */
	portPop(obj)

	/*
	* Remove ourself from tock/tick,
	* we will be re-added in link if there
	* is some need for us to be in there
	 */
	EventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
	EventQueueAdd(obj, QUEUE_TYPE_TICK, true)
}

/* Add port to correct alias, increment */
func portAlias(obj *ObjData, port int, ptype uint8) {
	defer reportPanic("portAlias")
	if obj.Ports[port].link == nil {
		return
	}

	switch ptype {
	case PORT_IN:
		obj.inputs = append(obj.inputs, &obj.Ports[port])
		obj.numIn++
	case PORT_OUT:
		obj.outputs = append(obj.outputs, &obj.Ports[port])
		obj.numOut++
	case PORT_FIN:
		obj.fuelIn = append(obj.fuelIn, &obj.Ports[port])
		obj.numFIn++
	case PORT_FOUT:
		obj.fuelOut = append(obj.fuelOut, &obj.Ports[port])
		obj.numFOut++
	}

	/* Fix trapped materials */
	for p, port := range obj.Ports {

		if port.Buf == nil {
			continue
		}

		if port.Buf.Amount == 0 {
			continue
		}

		var good bool
		var canFix bool
		switch port.Type {
		case PORT_IN:
			if obj.numIn > 0 {
				for op := range obj.inputs {
					if obj.inputs[op] == &obj.Ports[p] {
						/* Don't need to reprocess, port is alive */
						good = true
						break
					}
				}
				if !good {
					canFix = true
				}
			}
		case PORT_OUT:
			if obj.numOut > 0 {
				for op := range obj.outputs {
					if obj.outputs[op] == &obj.Ports[p] {
						/* Don't need to reprocess, port is alive */
						good = true
						break
					}
				}
				if !good {
					canFix = true
				}
			}
		case PORT_FIN:
			if obj.numFIn > 0 {
				for op := range obj.fuelIn {
					if obj.fuelIn[op] == &obj.Ports[p] {
						/* Don't need to reprocess, port is alive */
						good = true
						break
					}
				}
				if !good {
					canFix = true
				}
			}
		case PORT_FOUT:
			if obj.numFOut > 0 {
				for op := range obj.fuelOut {
					if obj.fuelOut[op] == &obj.Ports[p] {
						/* Don't need to reprocess, port is alive */
						good = true
						break
					}
				}
				if !good {
					canFix = true
				}
			}
		}

		if !good && canFix {
			fixed := false
			switch port.Type {
			case PORT_IN:
				if obj.inputs[0].Buf.Amount == 0 {
					/* Swap pointers */
					obj.inputs[0].Buf, obj.Ports[p].Buf = obj.Ports[p].Buf, obj.inputs[0].Buf
					fixed = true
				}
			case PORT_OUT:
				if obj.outputs[0].Buf.Amount == 0 {
					/* Swap pointers */
					obj.outputs[0].Buf, obj.Ports[p].Buf = obj.Ports[p].Buf, obj.outputs[0].Buf
					fixed = true
				}
			case PORT_FIN:
				if obj.fuelIn[0].Buf.Amount == 0 {
					/* Swap pointers */
					obj.fuelIn[0].Buf, obj.Ports[p].Buf = obj.Ports[p].Buf, obj.fuelIn[0].Buf
					fixed = true
				}
			case PORT_FOUT:
				if obj.fuelOut[0].Buf.Amount == 0 {
					/* Swap pointers */
					obj.fuelOut[0].Buf, obj.Ports[p].Buf = obj.Ports[p].Buf, obj.fuelOut[0].Buf
					fixed = true
				}
			}
			if fixed {
				DoLog(true, "Fixed orphaned material in object ports.")
			}
		}
	}
}

/*
 * Remove a port from aliases
 * Currently very lazy, but simple
 */
func portPop(obj *ObjData) {
	defer reportPanic("portPop")
	obj.outputs = nil
	obj.numOut = 0

	obj.inputs = nil
	obj.numIn = 0

	obj.fuelIn = nil
	obj.numFIn = 0

	obj.fuelOut = nil
	obj.numFOut = 0

	for p, port := range obj.Ports {
		portAlias(obj, p, port.Type)
	}
}
