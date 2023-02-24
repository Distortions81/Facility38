package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
)

/* Link to output in (dir) */
func LinkObj(b *world.BuildingData) {

	/* Link inputs if we have any */
	for p, port := range b.SubObj.Ports {

		/* Make sure port is unoccupied */
		if port.Obj != nil {
			continue
		}

		/* Get subobj world position */
		lpos := util.AddXY(b.Obj.Pos, b.SubObj.SubPos)
		neigh := util.GetNeighborObj(lpos, port.Dir)

		/* We found one*/
		if neigh == nil {
			continue
		}

		/* Neighbor port is available */
		for n, np := range neigh.SubObj.Ports {
			if np.Dir == util.ReverseDirection(port.Dir) {
				/* Assign on both sides */
				neigh.SubObj.Ports[n].Obj = b.Obj
				b.SubObj.Ports[p].Obj = neigh.Obj

				neigh.SubObj.Ports[n].Link = &b.SubObj.Ports[p]
				b.SubObj.Ports[p].Link = &neigh.SubObj.Ports[n]
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
