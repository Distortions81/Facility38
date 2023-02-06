package objects

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/util"
)

/* Link to output in (dir) */
func linkOut(pos glob.XY, obj *glob.ObjData, dir uint8) {

	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have outputs */
	if !obj.TypeP.HasMatOutput {
		//cwlog.DoLog("(%v: %v, %v) linkOut: we do not have any outputs", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	/* Look for object in output direction */
	neigh, _ := util.GetNeighborObj(obj, pos, dir)

	/* Did we find and obj? */
	if neigh == nil {
		//cwlog.DoLog("(%v: %v, %v) linkOut: Rejected nil neighbor: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		return
	}
	npos := util.CenterXY(neigh.Pos)

	/* Does it have inputs? */
	if neigh.TypeP.HasMatInput == 0 {
		//cwlog.DoLog("(%v: %v, %v) linkOut: Rejected: neighbor has no inputs: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		return
	}
	/* Do they have an output? */
	if neigh.TypeP.HasMatOutput {
		/* Are we trying to connect from that direction? */
		if neigh.TypeP.Direction == util.ReverseDirection(dir) {
			cwlog.DoLog("(%v: %v, %v) linkOut: Rejected: neighbor outputs this direction: %v: %v: (%v,%v)",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(util.ReverseDirection(dir)), neigh.TypeP.Name, npos.X, npos.Y)
			return
		}
	}

	/* If we have an output already, unlink it */
	if obj.OutputObj != nil {
		/* Unlink OLD output specifically */
		unlinkOut(obj.OutputObj)
		cwlog.DoLog("(%v: %v, %v) linkOut: removing our old output: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(obj.Direction))
	} else {
		obj.OutputBuffer = &glob.MatData{}
		cwlog.DoLog("(%v: %v, %v) linkOut: init our output buffer.", obj.TypeP.Name, ppos.X, ppos.Y)
	}

	/* Make sure the object has an input initialized */
	if neigh.InputBuffer[util.ReverseDirection(dir)] != nil {
		neigh.InputBuffer[util.ReverseDirection(dir)] = &glob.MatData{}
		cwlog.DoLog("(%v: %v, %v) linkOut: init neighbor input: %v: %v: (%v,%v)",
			obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
	}

	/* Mark target as our output */
	obj.OutputObj = neigh

	/* Put ourself in target's input list */
	neigh.InputObjs[util.ReverseDirection(dir)] = obj

	cwlog.DoLog("(%v: %v, %v) linkOut: Linked: %v: %v: (%v,%v)",
		obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), obj.OutputObj.TypeP.Name, npos.X, npos.Y)
}

/* Find and link inputs, set ourself to OutputObj of found objects */
func linkIn(pos glob.XY, obj *glob.ObjData, newdir uint8) {
	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have inputs */
	if obj.TypeP.HasMatInput == 0 {
		//cwlog.DoLog("(%v: %v, %v) linkIn: we have no inputs.", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	var dir uint8
	for dir = consts.DIR_NORTH; dir < consts.DIR_MAX; dir++ {

		/* Don't try to connect an input the same direction as our future output */
		/* If there is an input there, remove it */
		if obj.TypeP.HasMatOutput && dir == newdir {
			unlinkInput(obj, dir)
			cwlog.DoLog("(%v: %v, %v) linkIn: unlinking input that is in direction of our new output: %v",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Look for neighbor object */
		neigh, _ := util.GetNeighborObj(obj, pos, dir)

		/* Did we find an object? */
		if neigh == nil {
			//cwlog.DoLog("(%v: %v, %v) linkIn: nil neighbor: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}
		npos := util.CenterXY(neigh.Pos)

		/* Does it have an output? */
		if !neigh.TypeP.HasMatOutput {
			//cwlog.DoLog("(%v: %v, %v) linkIn: neighbor has no outputs: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Is the output unoccupied? */
		if neigh.OutputObj != nil {
			/* Is it us? */
			if neigh.OutputObj != obj {
				cwlog.DoLog("(%v: %v, %v) linkIn: neigbor output is occupied: %v: %v: (%v,%v)",
					obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
				continue
			}
		}

		/* Is the output in our direction? */
		if neigh.Direction != util.ReverseDirection(dir) {
			cwlog.DoLog("(%v: %v, %v) linkIn: neighbor output is not in our direction: %v: %v: (%v,%v)",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
			continue
		}

		/* Unlink old input from this direction if it exists */
		unlinkInput(obj, dir)

		/* Make sure they have an output initalized */
		if neigh.OutputBuffer == nil {
			neigh.OutputBuffer = &glob.MatData{}
			cwlog.DoLog("(%v: %v, %v) linkIn: initializing neighbor output: %v: %v: (%v,%v)",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
		}

		/* Make sure we have a input initalized */
		if obj.InputBuffer[dir] == nil {
			obj.InputBuffer[dir] = &glob.MatData{}
			cwlog.DoLog("(%v: %v, %v) linkIn: initializing our input : %v",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		}

		/* Set ourself as their output */
		linkOut(neigh.Pos, neigh, neigh.Direction)

		/* Record who is on this input */
		obj.InputObjs[util.ReverseDirection(dir)] = neigh
		obj.InputCount++

		cwlog.DoLog("(%v: %v, %v) linkIn: linked: %v: %v: (%v,%v)",
			obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
	}

}

/* Link inputs and outputs, with output direction (newdir) */
func LinkObj(pos glob.XY, obj *glob.ObjData, newdir uint8) {
	linkIn(pos, obj, newdir)
	linkOut(pos, obj, newdir)
}
