package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
)

/* Link to output in (dir) */
func LinkObj(b *world.BuildingData) {

	/* Attempt to link ports */
	for p, port := range b.Obj.Ports {
		fmt.Println("MEEP LINK")

		/* Make sure port is unoccupied */
		if port.Obj != nil {
			continue
		}

		/* Get world obj sub-position */
		lpos := util.AddXY(b.Obj.Pos, b.SubPos)
		neighb := util.GetNeighborObj(lpos, port.Dir)

		/* We found one*/
		if neighb == nil {
			continue
		}

		fmt.Println("MEEP FOUND NEIGH")

		/* Neighbor port is available */
		for n, np := range neighb.Obj.Ports {
			fmt.Println(np.Type, np.Dir)
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

				fmt.Println("MEEP aliased")
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

		portUnAlias(obj, dir)
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

	fmt.Println(obj.TypeP.Name, port)
}

func portUnAlias(obj *world.ObjData, port int) {
	obj.Inputs = nil
	obj.NumIn = 0
	obj.Outputs = nil
	obj.NumOut = 0
	obj.FuelIn = nil
	obj.NumFIn = 0
	obj.FuelOut = nil
	obj.NumFOut = 0
}
