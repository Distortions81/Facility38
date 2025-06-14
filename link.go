package main

import (
	"fmt"
	"sync"
)

var linkLock sync.Mutex

func linkObj(from XY, b *buildingData) {

	b.obj.selected = true

	defer func() {
		b.obj.selected = false
	}()

	/* multi-tile object relink */
	if b.obj.Unique.typeP.multiTile {
		for _, subObj := range b.obj.Unique.typeP.subObjs {
			subPos := GetSubPos(b.obj.Pos, subObj)
			linkSingleObj(subPos, b)
		}
	} else {
		/* Standard relink */
		linkSingleObj(b.obj.Pos, b)
	}
}

/* Link to output in (dir) */
func linkSingleObj(from XY, b *buildingData) {
	defer reportPanic("linkObj")

	linkLock.Lock()
	defer linkLock.Unlock()

	objCD(b, fmt.Sprintf("Facing: %v", dirToName(b.obj.Dir)))
	b.obj.LastInput = 0
	b.obj.LastOutput = 0

	/* Attempt to link ports */
	for p, port := range b.obj.Ports {
		/* Make sure port is unoccupied */
		if port.obj != nil {
			objCD(b, fmt.Sprintf("Port Occupied: %v", dirToName(port.Dir)))
			continue
		}

		var neighbor *buildingData
		if port.Dir == DIR_ANY {
			var testPort uint8
			for testPort = DIR_NORTH; testPort <= DIR_WEST; testPort++ {
				//doLog(true, "Looking in all directions: "+dirToName(testPort))

				neighbor = getNeighborObj(from, testPort)
				if neighbor != nil && neighbor.obj != nil {
					if neighbor.obj.Pos != b.obj.Pos {
						//doLog(true, "found")
						break
					}
				}
			}
			if neighbor == nil {
				continue
			}
		} else {
			neighbor = getNeighborObj(from, port.Dir)
		}

		/* We found one*/
		if neighbor == nil {
			objCD(b, fmt.Sprintf("No neighbor: %v", dirToName(port.Dir)))
			continue
		}

		if neighbor.obj.Pos == b.obj.Pos {
			objCD(b, fmt.Sprintf("Ignoring link to self: %v", dirToName(port.Dir)))
			continue
		}

		if infoLine {
			doLog(true, dirToName(port.Dir))
		}

		for n, np := range neighbor.obj.Ports {

			/* Neighbor port is available */
			if np.obj != nil {
				objCD(b, fmt.Sprintf("Port occupied: %v", dirToName(port.Dir)))
				continue
			}

			/* Port is in correct direction */
			if np.Dir != DIR_MAX &&
				np.Dir == reverseDirection(port.Dir) ||
				np.Dir == DIR_ANY ||
				port.Dir == DIR_ANY {

				/* Port is of correct type */
				if port.Type != reverseType(np.Type) {
					objCD(b, fmt.Sprintf("Port incorrect type: %v", typeToName(port.Type)))
					continue
				}

				/* Normal objects can only link to loaders */
				if (b.obj.Unique.typeP.category == objCatGeneric &&
					neighbor.obj.Unique.typeP.category != objCatLoader) ||

					//Belts can only link to belts or loaders
					(b.obj.Unique.typeP.category == objCatBelt &&
						neighbor.obj.Unique.typeP.category != objCatLoader &&
						neighbor.obj.Unique.typeP.category != objCatBelt) {
					continue
				}

				/* Add link to objects */
				neighbor.obj.Ports[n].obj = b.obj
				b.obj.Ports[p].obj = neighbor.obj

				/* Add direct port links */
				neighbor.obj.Ports[n].link = &b.obj.Ports[p]
				b.obj.Ports[p].link = &neighbor.obj.Ports[n]

				if debugMode {
					oName := "none"
					if b.obj != nil {
						oName = fmt.Sprintf("%v: %v", neighbor.obj.Unique.typeP.name, posToString(neighbor.pos))
					}
					objCD(b, fmt.Sprintf("Linked: Port-%v: ( %v %v ) to %v", p, dirToName(port.Dir), DirToArrow(port.Dir), oName))
				}

				portAlias(b.obj, p, port.Type)
				portAlias(neighbor.obj, n, np.Type)

				/* Run custom link code */
				if neighbor.obj.Unique.typeP.cLinkObj != nil {
					neighbor.obj.Unique.typeP.cLinkObj(neighbor.obj)
				} else {
					autoEvents(neighbor.obj)
				}

				/*
				 * Only allow one neighbor port to link to
				 * this port.
				 */
				break
			}
		}
		/* Run custom link code */
		if b.obj.Unique.typeP.cLinkObj != nil {
			b.obj.Unique.typeP.cLinkObj(b.obj)
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
			eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}
		if obj.Unique.typeP.hasOutputs && foundOutputs {
			eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
			eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}

		if obj.Unique.typeP.hasFIn && foundFIn {
			eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
		}
		if obj.Unique.typeP.hasFOut && foundFOut {
			eventQueueAdd(obj, QUEUE_TYPE_TICK, false)
			eventQueueAdd(obj, QUEUE_TYPE_TOCK, false)
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
		for pb, ourPort := range port.obj.Ports {
			if ourPort.obj == obj {
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
				if pObj.Unique.typeP.cLinkObj != nil {
					pObj.Unique.typeP.cLinkObj(pObj)
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
	eventQueueAdd(obj, QUEUE_TYPE_TOCK, true)
	eventQueueAdd(obj, QUEUE_TYPE_TICK, true)
}

/* Add port to correct alias, increment */
func portAlias(obj *ObjData, port int, pType uint8) {
	defer reportPanic("portAlias")
	if obj.Ports[port].link == nil {
		return
	}

	switch pType {
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
				doLog(true, "Fixed orphaned material in object ports.")
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
	obj.outputs = []*ObjPortData{}
	obj.numOut = 0

	obj.inputs = []*ObjPortData{}
	obj.numIn = 0

	obj.fuelIn = []*ObjPortData{}
	obj.numFIn = 0

	obj.fuelOut = []*ObjPortData{}
	obj.numFOut = 0

	for p, port := range obj.Ports {
		portAlias(obj, p, port.Type)
	}
}
