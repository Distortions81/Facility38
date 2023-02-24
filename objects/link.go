package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
)

/* Link to output in (dir) */
func LinkObj(b *world.BuildingData) {

	/* Link inputs if we have any */
	for p, port := range b.Obj.Ports {

		/* Make sure port is unoccupied */
		if port.Obj != nil {
			continue
		}

		/* Get world obj sub-position */
		lpos := util.AddXY(b.Obj.Pos, b.SubPos)
		neigh := util.GetNeighborObj(lpos, port.Dir)

		/* We found one*/
		if neigh == nil {
			continue
		}

		/* Neighbor port is available */
		for n, np := range neigh.Obj.Ports {
			if np.Dir == util.ReverseDirection(port.Dir) {
				/* Assign on both sides */
				neigh.Obj.Ports[n].Obj = b.Obj
				b.Obj.Ports[p].Obj = neigh.Obj

				neigh.Obj.Ports[n].Link = &b.Obj.Ports[p]
				b.Obj.Ports[p].Link = &neigh.Obj.Ports[n]
			}
		}
	}
}

/* UnlinkObj an object's (dir) input */
func UnlinkObj(obj *world.ObjData) {

	for dir, port := range obj.Ports {

		/* Change object port accounting */
		if port.PortDir == gv.PORT_INPUT {
			obj.NumInputs--
			if port.Obj != nil {
				port.Obj.NumOutputs--

				rObj := port.Obj
				rObj.Ports[util.ReverseDirection(uint8(dir))].Obj = nil

				obj.Ports[dir].Obj = nil
			}
		} else if port.PortDir == gv.PORT_OUTPUT {
			obj.NumOutputs++
			if port.Obj != nil {
				port.Obj.NumInputs--

				rObj := port.Obj
				rObj.Ports[util.ReverseDirection(uint8(dir))].Obj = nil

				obj.Ports[dir].Obj = nil
			}
		}
	}
}
