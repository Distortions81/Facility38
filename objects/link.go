package objects

import (
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
)

/* Link to output in (dir) */
func LinkObj(obj *glob.ObjData) {

	/* Check our ports */
	for p, port := range obj.Ports {

		/* Make sure it is empty */
		if port.Obj != nil {
			continue
		}

		/* Look for neighbor in that direction */
		neigh := util.GetNeighborObj(obj, uint8(p))

		/* We found one */
		if neigh == nil {
			continue
		}

		/* Port is in correct direction */
		if port.PortDir == neigh.Ports[util.ReverseDirection(uint8(p))].PortDir {
			continue
		}

		/* Port is available */
		if neigh.Ports[util.ReverseDirection(uint8(p))].Obj != nil {
			continue
		} else {
			/* Unlink old */
			obj.Ports[p].Obj = nil

			if port.PortDir == gv.PORT_INPUT {
				neigh.NumOutputs--
				obj.NumInputs--
			} else {
				neigh.NumInputs--
				obj.NumOutputs--
			}
		}

		/* Assign on both sides */
		neigh.Ports[util.ReverseDirection(uint8(p))].Obj = obj
		obj.Ports[p].Obj = neigh

		/* add to input/output counts */
		if port.PortDir == gv.PORT_INPUT {
			neigh.NumOutputs++
			obj.NumInputs++
		} else {
			neigh.NumInputs++
			obj.NumOutputs++
		}
	}
}
