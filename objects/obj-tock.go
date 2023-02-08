package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
)

func minerUpdate(obj *glob.ObjData) {
	/* Hard-coded for speed */
	if obj.Ports[obj.Dir].Buf.Amount == 0 {
		obj.Blocked = false

		obj.Ports[obj.Dir].Buf.Amount = obj.TypeP.MinerKGTock
		obj.Ports[obj.Dir].Buf.TypeP = *MatTypes[gv.MAT_COAL]
	} else {
		obj.Blocked = true
	}
}

func beltUpdate(obj *glob.ObjData) {

	/* Output is full, exit */
	if obj.Ports[obj.Dir].Buf.Amount != 0 {
		cwlog.DoLog("beltUpdate: Our output is full. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
		return
	}

	/* Find all inputs, round-robin send to output */
	for p, port := range obj.Ports {
		if port.PortDir != gv.PORT_INPUT {
			continue
		}
		if obj.NumInputs > 1 {
			if uint8(p) == obj.LastUsedInput {
				cwlog.DoLog("beltUpdate: Skipping previously used input.%v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
				continue
			}
			obj.LastUsedInput = uint8(p)
		}
		if port.Buf.Amount == 0 {
			cwlog.DoLog("beltUpdate: Our input is empty. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		} else {
			obj.Ports[obj.Dir].Buf.Amount = port.Buf.Amount
			obj.Ports[obj.Dir].Buf.TypeP = port.Buf.TypeP
			obj.Ports[p].Buf.Amount = 0
			break
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
	for p, port := range obj.Ports {
		if port.PortDir != gv.PORT_INPUT {
			cwlog.DoLog("tickObj: Our port is not an input. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		}

		if port.Buf.Amount == 0 {
			cwlog.DoLog("tickObj: Input is empty. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		}

		if obj.KGHeld+port.Buf.Amount > obj.TypeP.CapacityKG {
			cwlog.DoLog("boxUpdate: Object is full %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		}
		if obj.Contents[port.Buf.TypeP.TypeI] == nil {
			obj.Contents[port.Buf.TypeP.TypeI] = &glob.MatData{}
		}
		obj.Contents[port.Buf.TypeP.TypeI].Amount += obj.Ports[p].Buf.Amount
		obj.Contents[port.Buf.TypeP.TypeI].TypeP = obj.Ports[p].Buf.TypeP
		obj.KGHeld += port.Buf.Amount
		obj.Ports[p].Buf.Amount = 0

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
