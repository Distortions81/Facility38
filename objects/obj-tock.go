package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
)

func minerUpdate(obj *glob.ObjData) {
	/* Hard-coded for speed */
	if obj.Ports[obj.Dir].Buf.Amount == 0 {
		input := obj.TypeP.MinerKGTock

		obj.Ports[obj.Dir].Buf.Amount = input
		obj.Ports[obj.Dir].Buf.TypeP = *MatTypes[gv.MAT_COAL]
	}
}

func beltUpdate(obj *glob.ObjData) {
	/* No outputs */
	if obj.NumOutputs == 0 {
		return
	}

	/* Output is full, exit */
	if obj.Ports[obj.Dir].Buf.Amount > 0 {
		return
	}

	/* Find all inputs, round-robin send to output */
	for p, port := range obj.Ports {
		if port.PortDir == gv.PORT_INPUT {
			if obj.NumInputs > 1 {
				if uint8(p) == obj.LastUsedInput {
					continue
				}
				obj.LastUsedInput = uint8(p)
			}
			if port.Buf.Amount > 0 {
				obj.Ports[obj.Dir].Buf.Amount = port.Buf.Amount
				obj.Ports[obj.Dir].Buf.TypeP = port.Buf.TypeP
				port.Buf.Amount = 0
				break
			}
		}
	}
}

func splitterUpdate(obj *glob.ObjData) {
	if obj.NumOutputs <= 0 {
		return
	}

	for _, port := range obj.Ports {
		if port.PortDir == gv.PORT_INPUT {
			if port.Buf.Amount > 0 {
				for _, oport := range obj.Ports {
					if oport.PortDir == gv.PORT_OUTPUT {
						if oport.Buf.Amount <= 0 {
							oport.Buf.Amount = port.Buf.Amount
							oport.Buf.TypeP = port.Buf.TypeP
							port.Buf.Amount = 0
						}
					}
				}
			}
		}
	}
}

func boxUpdate(obj *glob.ObjData) {
	if obj.NumInputs <= 0 {
		return
	}

	for _, port := range obj.Ports {
		if port.Buf.Amount > 0 {
			if obj.KGHeld+port.Buf.Amount > obj.TypeP.CapacityKG {
				cwlog.DoLog("%v: Object is full.", obj.TypeP.Name)
				continue
			}
			obj.Contents[port.Buf.TypeP.TypeI].Amount = port.Buf.Amount
			obj.Contents[port.Buf.TypeP.TypeI].TypeP = port.Buf.TypeP
			obj.KGHeld += port.Buf.Amount
			port.Buf.Amount = 0
		}
	}
}

func smelterUpdate(o *glob.ObjData) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func ironCasterUpdate(o *glob.ObjData) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func steamEngineUpdate(o *glob.ObjData) {
}
