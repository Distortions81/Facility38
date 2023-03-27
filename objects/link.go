package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"sync"
)

/* Link to output in (dir) */
var linkLock sync.Mutex

func LinkObj(from world.XY, b *world.BuildingData) {

	linkLock.Lock()
	defer linkLock.Unlock()

	util.ObjCD(b, fmt.Sprintf("Facing: %v", util.DirToName(b.Obj.Dir)))
	b.Obj.LastInput = 0
	b.Obj.LastOutput = 0

	/* Attempt to link ports */
	for p, port := range b.Obj.Ports {

		/* Make sure port is unoccupied */
		if port.Obj != nil {
			util.ObjCD(b, fmt.Sprintf("Port Occupied: %v", util.DirToName(port.Dir)))
			continue
		}

		neighb := util.GetNeighborObj(from, port.Dir)

		/* We found one*/
		if neighb == nil {
			util.ObjCD(b, fmt.Sprintf("No neighbor: %v", util.DirToName(port.Dir)))
			continue
		}

		if neighb.Obj.Pos == b.Obj.Pos {
			//util.ObjCD(b, fmt.Sprintf("Ignoring link to self: %v", util.DirToName(port.Dir)))
			continue
		}

		for n, np := range neighb.Obj.Ports {

			/* Neighbor port is available */
			if np.Obj != nil {
				continue
			}

			/* Port is in correct direction */
			if np.Dir == util.ReverseDirection(port.Dir) ||
				np.Dir == gv.DIR_ANY || port.Dir == gv.DIR_ANY {

				/* Port is of correct type */
				if port.Type != util.ReverseType(np.Type) {
					util.ObjCD(b, fmt.Sprintf("Port incorrect type: %v", util.DirToName(port.Dir)))
					continue
				}

				/* Normal objects can only link to loaders */
				if (b.Obj.TypeP.Category == gv.ObjCatGeneric &&
					neighb.Obj.TypeP.Category != gv.ObjCatLoader) ||
					(neighb.Obj.TypeP.Category == gv.ObjCatGeneric &&
						b.Obj.TypeP.Category != gv.ObjCatLoader) {
					continue
				}

				/* Add link to objects */
				neighb.Obj.Ports[n].Obj = b.Obj
				b.Obj.Ports[p].Obj = neighb.Obj

				/* Add direct port links */
				neighb.Obj.Ports[n].Link = &b.Obj.Ports[p]
				b.Obj.Ports[p].Link = &neighb.Obj.Ports[n]

				if gv.Debug {
					oName := "none"
					if b.Obj != nil {
						oName = fmt.Sprintf("%v: %v", neighb.Obj.TypeP.Name, util.PosToString(neighb.Pos))
					}
					util.ObjCD(b, fmt.Sprintf("Linked: Port-%v: ( %v %v ) to %v", p, util.DirToName(port.Dir), util.DirToArrow(port.Dir), oName))
				}

				portAlias(b.Obj, p, port.Type)
				portAlias(neighb.Obj, n, np.Type)

				/* Run custom link code */
				if neighb.Obj.TypeP.LinkObj != nil {
					neighb.Obj.TypeP.LinkObj(neighb.Obj)
				} else {
					AutoEvents(neighb.Obj)
				}
			}
		}
		/* Run custom link code */
		if b.Obj.TypeP.LinkObj != nil {
			b.Obj.TypeP.LinkObj(b.Obj)
		} else {
			AutoEvents(b.Obj)
		}

	}

}

/* Add/Remove tick/tock events as needed */
func AutoEvents(obj *world.ObjData) {

	/* Add to tock/tick lists */
	var foundOutputs, foundInputs, foundFOut, foundFIn bool
	if obj.NumOut > 0 {
		foundOutputs = true
	}
	if obj.NumIn > 0 {
		foundInputs = true
	}
	if obj.NumFIn > 0 {
		foundFIn = true
	}
	if obj.NumFOut > 0 {
		foundFOut = true
	}

	/* If we have inputs and outputs object needs, add to tock list */
	if obj.TypeP.UpdateObj != nil {

		if obj.TypeP.HasInputs && foundInputs {
			EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		}
		if obj.TypeP.HasOutputs && foundOutputs {
			EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
			EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		}

		if obj.TypeP.HasFIn && foundFIn {
			EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		}
		if obj.TypeP.HasFOut && foundFOut {
			EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
			EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
		}
	}
}

/* UnlinkObj an object */
func UnlinkObj(obj *world.ObjData) {

	linkLock.Lock()
	defer linkLock.Unlock()

	/* Reset last input var */
	obj.LastInput = 0
	obj.LastOutput = 0

	for p, port := range obj.Ports {
		/* No obj, skip */
		if port.Obj == nil {
			continue
		}

		/* Delete ourselves from others */
		for pb, portb := range port.Obj.Ports {
			if portb.Obj == obj {
				pObj := port.Obj

				/* Reset last port to avoid hitting invalid one */
				if port.Type == gv.PORT_IN {
					obj.LastInput = 0
				} else {
					port.Obj.LastInput = 0
				}

				/* Clean up port aliases */
				pObj.Ports[pb].Link = nil
				pObj.Ports[pb].Obj = nil

				portPop(pObj)
				if pObj.TypeP.LinkObj != nil {
					pObj.TypeP.LinkObj(pObj)
				} else {
					AutoEvents(pObj)
				}
			}
		}
		portPop(port.Obj)

		/* Break links */
		obj.Ports[p].Link = nil
		obj.Ports[p].Obj = nil
	}

	/* Murder all the port aliases */
	portPop(obj)

	/*
	* Remove ourself from tock/tick,
	* we will be re-added in link if there
	* is some need for us to be in there
	 */
	EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)
	EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
}

/* Add port to correct alias, increment */
func portAlias(obj *world.ObjData, port int, ptype uint8) {
	if obj.Ports[port].Link == nil {
		return
	}

	switch ptype {
	case gv.PORT_IN:
		obj.Inputs = append(obj.Inputs, &obj.Ports[port])
		obj.NumIn++
	case gv.PORT_OUT:
		obj.Outputs = append(obj.Outputs, &obj.Ports[port])
		obj.NumOut++
	case gv.PORT_FIN:
		obj.FuelIn = append(obj.FuelIn, &obj.Ports[port])
		obj.NumFIn++
	case gv.PORT_FOUT:
		obj.FuelOut = append(obj.FuelOut, &obj.Ports[port])
		obj.NumFOut++
	}
}

/*
 * Remove a port from aliases
 * Currently very lazy, but simple
 */
func portPop(obj *world.ObjData) {
	obj.Outputs = nil
	obj.NumOut = 0

	obj.Inputs = nil
	obj.NumIn = 0

	obj.FuelIn = nil
	obj.NumFIn = 0

	obj.FuelOut = nil
	obj.NumFOut = 0

	for p, port := range obj.Ports {
		portAlias(obj, p, port.Type)
	}
}
