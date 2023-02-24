package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
)

/* Link to output in (dir) */
func LinkObj(b *world.BuildingData) {

	/* Link inputs if we have any */
	for p, port := range b.SubObj.InPorts {

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
		for n, np := range neigh.SubObj.OutPorts {
			if np.Dir == util.ReverseDirection(port.Dir) {
				/* Assign on both sides */
				neigh.SubObj.OutPorts[n].Obj = b.Obj
				b.SubObj.InPorts[p].Obj = neigh.Obj
			}
		}
	}

	/* Link outputs if we have any */
	for p, port := range b.SubObj.OutPorts {

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
		for n, np := range neigh.SubObj.InPorts {
			if np.Dir == util.ReverseDirection(port.Dir) {
				/* Assign on both sides */
				neigh.SubObj.InPorts[n].Obj = b.Obj
				b.SubObj.OutPorts[p].Obj = neigh.Obj
			}
		}
	}

	/* Link fuel outputs if we have any */
	for p, port := range b.SubObj.FuelOut {

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
		for n, np := range neigh.SubObj.FuelIn {
			if np.Dir == util.ReverseDirection(port.Dir) {
				/* Assign on both sides */
				neigh.SubObj.FuelIn[n].Obj = b.Obj
				b.SubObj.FuelOut[p].Obj = neigh.Obj
			}
		}
	}

	/* Link fuel inputs if we have any */
	for p, port := range b.SubObj.FuelIn {

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
		for n, np := range neigh.SubObj.FuelOut {
			if np.Dir == util.ReverseDirection(port.Dir) {
				/* Assign on both sides */
				neigh.SubObj.FuelOut[n].Obj = b.Obj
				b.SubObj.FuelIn[p].Obj = neigh.Obj
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
