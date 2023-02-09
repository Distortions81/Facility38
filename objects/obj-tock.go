package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/noise"
	"GameTest/util"
	"math/rand"
)

func minerUpdate(obj *glob.ObjData) {

	obj.TickCount++
	if obj.TickCount < obj.TypeP.Interval {
		return
	}
	obj.TickCount = 0

	/* Hard-coded for speed */
	if obj.Ports[obj.Dir].Buf.Amount == 0 {
		obj.Blocked = false

		var matsFound [noise.NumNoiseTypes]float64
		var matsFoundT [noise.NumNoiseTypes]uint8
		numTypesFound := 0

		/* TODO: Optimize, only run this once when placed */
		for p := 1; p < 5; p++ {
			h := 1.0 - (noise.NoiseMap(float64(obj.Pos.X), float64(obj.Pos.Y), p) * 2)

			if h > 0.5 {
				//fmt.Println(h, obj.Pos)
				matsFound[numTypesFound] = h
				matsFoundT[numTypesFound] = uint8(p)
				numTypesFound++
			}
		}

		if numTypesFound > 0 {
			pick := rand.Intn(numTypesFound)

			obj.Ports[obj.Dir].Buf.Amount = obj.TypeP.MinerKGTock * matsFound[pick]
			obj.Ports[obj.Dir].Buf.TypeP = *MatTypes[matsFoundT[pick]]
			obj.Ports[obj.Dir].Buf.Rot = uint8(rand.Intn(3))
		} else {
			obj.Blocked = true
		}
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
			if obj.LastUsedInput < obj.NumInputs {
				obj.LastUsedInput++
			} else {
				obj.LastUsedInput = 0
			}
		}
		if port.Buf.Amount == 0 {
			cwlog.DoLog("beltUpdate: Our input is empty. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		} else {
			obj.Ports[obj.Dir].Buf.Amount = port.Buf.Amount
			obj.Ports[obj.Dir].Buf.TypeP = port.Buf.TypeP
			obj.Ports[obj.Dir].Buf.Rot = port.Buf.Rot
			obj.Ports[p].Buf.Amount = 0
			break
		}

	}
}

func splitterUpdate(obj *glob.ObjData) {
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
			obj.Ports[obj.Dir].Buf.Rot = port.Buf.Rot
			obj.Ports[p].Buf.Amount = 0
			break
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

func smelterUpdate(obj *glob.ObjData) {

	/* Find all inputs, round-robin send to output */
	for p, port := range obj.Ports {

		if port.PortDir == gv.PORT_INPUT {

			/* Are we full? */
			if obj.KGHeld+port.Buf.Amount > obj.TypeP.CapacityKG {
				cwlog.DoLog("Smelter full")
				continue
			}

			/* If this is fuel or ore, take it */
			if port.Buf.TypeP.TypeI == gv.MAT_COAL ||
				port.Buf.TypeP.IsOre {

				if obj.Contents[port.Buf.TypeP.TypeI] == nil {
					obj.Contents[port.Buf.TypeP.TypeI] = &glob.MatData{}
					obj.Contents[port.Buf.TypeP.TypeI].TypeP = port.Buf.TypeP
				}

				obj.Contents[port.Buf.TypeP.TypeI].Amount += port.Buf.Amount

				obj.Ports[p].Buf.Amount = 0
				obj.KGHeld += port.Buf.Amount

				cwlog.DoLog("Accepted: %v", port.Buf.TypeP.Name)
			}
		} else {

			/* Output is full, exit */
			if port.Buf.Amount != 0 {
				cwlog.DoLog("smelterUpdate: Our output is blocked. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
				continue
			}

			obj.TickCount++
			if obj.TickCount >= obj.TypeP.Interval {
				typeCount := 0

				/* Smelt stuff */
				//if obj.Contents[gv.MAT_COAL] != nil && obj.Contents[gv.MAT_COAL].Amount >= 0.0 {
				for c, cont := range obj.Contents {
					if cont == nil {
						continue
					}

					if cont.TypeP.IsOre && cont.Amount >= 0.75 {
						typeCount++
						//obj.Contents[gv.MAT_COAL].Amount -= 0.125
						obj.Contents[c].Amount -= 0.75
						obj.KGHeld -= 0.75

						obj.Ports[p].Buf.Amount = 0.75
						if typeCount == 1 {
							obj.Ports[p].Buf.TypeP = *MatTypes[cont.TypeP.Result]
						}
						obj.Ports[p].Buf.Rot = port.Buf.Rot
					}
				}
				//}
				obj.TickCount = 0
			}
		}

	}
}

func ironCasterUpdate(o *glob.ObjData) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func steamEngineUpdate(o *glob.ObjData) {
}
