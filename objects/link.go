package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
)

/* Link to output in (dir) */
func LinkObj(obj *glob.ObjData) {
	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("LinkObj: %v (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

	/* Check our ports */
	for p, port := range obj.Ports {

		if obj.Ports[p] == nil {
			obj.Ports[p] = &glob.ObjPortData{}
			continue
		}

		/* Make sure our port is empty */
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
			//cwlog.DoLog("LinkObj: %v: %v (%v,%v): PortDir is wrong.", obj.TypeP.Name, util.DirToName(uint8(p)), oPos.X, oPos.Y)
			continue
		}

		/* Port is available */
		if neigh.Ports[util.ReverseDirection(uint8(p))].Obj != nil {
			cwlog.DoLog("LinkObj: %v: %v (%v,%v): Their port is in use.", obj.TypeP.Name, util.DirToName(uint8(p)), oPos.X, oPos.Y)
			continue
		}

		/* Assign on both sides */
		neigh.Ports[util.ReverseDirection(uint8(p))].Obj = obj
		obj.Ports[p].Obj = neigh

		cwlog.DoLog("LinkObj: %v: %v (%v,%v)", obj.TypeP.Name, util.DirToName(uint8(p)), oPos.X, oPos.Y)

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
