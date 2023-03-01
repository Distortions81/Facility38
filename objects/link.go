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
			}
		}
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

		obj.Ports[p].Link = nil
		obj.Ports[p].Obj = nil
	}
}

func portAlias(obj *world.ObjData, port int, ptype uint8) {
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
