package objects

import (
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
		lpos := util.AddXY(b.Obj.Pos, b.SubPos)
		neigh := util.GetNeighborObj(lpos, port.Dir)

		/* We found one*/
		if neigh == nil {
			continue
		}

		/* Neighbor port is available */
		for n, np := range neigh.Obj.Ports {
			/* Port is in correct direction */
			if np.Dir == util.ReverseDirection(port.Dir) &&
				/* Port is of correct type */
				np.Type == util.ReverseType(np.Type) {

				/* Assign on both sides */
				/* Add link to objects */
				neigh.Obj.Ports[n].Obj = b.Obj
				b.Obj.Ports[p].Obj = neigh.Obj

				/* Add direct port links */
				neigh.Obj.Ports[n].Link = &b.Obj.Ports[p]
				b.Obj.Ports[p].Link = &neigh.Obj.Ports[n]
			}
		}
	}
}

/* UnlinkObj an object's (dir) input */
func UnlinkObj(obj *world.ObjData) {

	for dir, port := range obj.Ports {
		/* No obj, skip */
		if port.Obj == nil {
			continue
		}
		port.Obj.Ports[util.ReverseDirection(uint8(dir))].Link = nil
		obj.Ports[dir].Link = nil

		port.Obj.Ports[util.ReverseDirection(uint8(dir))].Obj = nil
		obj.Ports[dir].Obj = nil
	}
}
