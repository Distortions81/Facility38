package objects

import (
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
)

func LinkObj(obj *glob.ObjData) {
	linkObjDir(obj, gv.PORT_INPUT)
	linkObjDir(obj, gv.PORT_OUTPUT)
}

/* Link to output in (dir) */
func linkObjDir(obj *glob.ObjData, portDir uint8) {

	/* Check our ports */
	for p, port := range obj.Ports {

		/* Port in correct direction */
		if port.PortDir != portDir {
			continue
		}

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
		if neigh.TypeP.Ports[p] != util.ReversePort(portDir) {
			continue
		}

		/* Port is available */
		if neigh.Ports[p].Obj != nil {
			continue
		} else {
			/* Unlink old */
			neigh.Ports[p].Obj = nil
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
		neigh.Ports[p].Obj = obj
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
