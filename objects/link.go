package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
)

/* Link to output in (dir) */
func LinkObj(b *world.BuildingData) {

	/* Attempt to link ports */
	for p, port := range b.Obj.Ports {

		/* Make sure port is unoccupied */
		if port.Obj != nil {
			continue
		}

		/* Get world obj sub-position */
		lpos := util.AddXY(b.SubPos, b.Obj.Pos)
		neighb := util.GetNeighborObj(lpos, port.Dir)

		/* We found one*/
		if neighb == nil {
			continue
		}

		/* Neighbor port is available */
		for n, np := range neighb.Obj.Ports {
			/* Port is in correct direction */
			if np.Dir == util.ReverseDirection(port.Dir) &&
				/* Port is of correct type */
				port.Type == util.ReverseType(np.Type) {

				/* Add link to objects */
				neighb.Obj.Ports[n].Obj = b.Obj
				b.Obj.Ports[p].Obj = neighb.Obj

				/* Add direct port links */
				neighb.Obj.Ports[n].Link = &b.Obj.Ports[p]
				b.Obj.Ports[p].Link = &neighb.Obj.Ports[n]

				portAlias(b.Obj, p, port.Type)
				portAlias(neighb.Obj, n, np.Type)

				AutoEvents(neighb.Obj)
			}
		}

	}

	AutoEvents(b.Obj)

}

/* Add/Remove tick/tock events as needed */
func AutoEvents(obj *world.ObjData) {

	/* Add to tock/tick lists */
	var foundOutputs, foundInputs bool
	if obj.NumOut > 0 || obj.NumFOut > 0 {
		foundOutputs = true
	}
	if obj.NumIn > 0 || obj.NumFIn > 0 {
		foundInputs = true
	}

	/* If we have inputs and outputs object needs, add to tock list */
	if obj.TypeP.UpdateObj != nil {
		if obj.TypeP.HasInputs && foundInputs {
			tockListAdd(obj)
		} else if obj.TypeP.HasOutputs && foundOutputs {
			tockListAdd(obj)
		} else {
			tocklistRemove(obj)
		}
	}
	/* Only add to tick list if object has an output */
	if obj.TypeP.HasOutputs && foundOutputs {
		ticklistAdd(obj)
	} else {
		ticklistRemove(obj)
	}
}

/* UnlinkObj an object */
func UnlinkObj(obj *world.ObjData) {

	for p, port := range obj.Ports {
		/* No obj, skip */
		if port.Obj == nil {
			continue
		}

		/* Delete ourselves from others */
		for pb, portb := range port.Obj.Ports {
			if portb.Obj == obj {
				port.Obj.Ports[pb].Link = nil
				port.Obj.Ports[pb].Obj = nil
			}
		}

		portPop(port.Obj)

		/* Break links */
		obj.Ports[p].Link = nil
		obj.Ports[p].Obj = nil
	}
	obj.Inputs = nil
	obj.Outputs = nil
	obj.FuelIn = nil
	obj.FuelOut = nil
	obj.NumIn = 0
	obj.NumOut = 0
	obj.NumFIn = 0
	obj.NumFOut = 0

	tocklistRemove(obj)
	ticklistRemove(obj)
}

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
