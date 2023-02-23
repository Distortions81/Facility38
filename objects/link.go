package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
)

/* Link to output in (dir) */
func LinkObj(obj *world.ObjData) {

	/* Check our ports */
	for p, port := range obj.Ports {

		if obj.Ports[p] == nil {
			continue
		}

		/* Make sure port is unoccupied */
		if port.Obj != nil {
			continue
		}

		/* Look for neighbor in direction */
		neigh := util.GetNeighborObj(obj, uint8(p))

		/* We found one*/
		if neigh == nil {
			continue
		}

		/* Port is in correct direction on their side */
		if port.PortDir == neigh.Ports[util.ReverseDirection(uint8(p))].PortDir {
			continue
		}

		/* Port is available */
		if neigh.Ports[util.ReverseDirection(uint8(p))].Obj != nil {
			continue
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
