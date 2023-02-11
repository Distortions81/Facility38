package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/noise"
	"GameTest/util"
	"math/rand"
)

func toggleOverlay() {
	if glob.ShowInfoLayer {
		glob.ShowInfoLayer = false
	} else {
		glob.ShowInfoLayer = true
	}
}

func InitMiner(obj *glob.ObjData) {
	if obj == nil {
		return
	}
	/* Init miner data if needed */
	if obj.MinerData == nil {
		obj.MinerData = &glob.MinerDataType{}
	}

	for p := 1; p < 5; p++ {
		h := 1.0 - (noise.NoiseMap(float64(obj.Pos.X), float64(obj.Pos.Y), p) * 2)

		if h > 0 {
			obj.MinerData.MatsFound[obj.MinerData.NumTypesFound] = h
			obj.MinerData.MatsFoundT[obj.MinerData.NumTypesFound] = noise.NoiseLayers[p].Type
			obj.MinerData.NumTypesFound++
		}
	}
}

func minerUpdate(obj *glob.ObjData) {

	/* Find all inputs, round-robin send to output */
	for p, port := range obj.Ports {
		if port == nil {
			continue
		}

		if port.PortDir == gv.PORT_INPUT {
			if port.Buf.TypeP == nil {
				continue
			}

			/* If this is fuel, take it */
			if port.Buf.TypeP.TypeI == gv.MAT_COAL {
				if obj.KGFuel+port.Buf.Amount > obj.TypeP.MaxFuelKG {
					continue
				}
				obj.KGFuel += port.Buf.Amount
				obj.Ports[p].Buf.Amount = 0
			}
		} else {

			/* Output is full, exit */
			if port.Buf.Amount != 0 {
				obj.Blocked = true
				obj.Active = false
				continue
			}
			obj.Blocked = false

			obj.TickCount++
			if obj.TickCount >= obj.TypeP.Interval {

				/* Mine stuff */
				if obj.KGFuel >= obj.TypeP.KgFuelEach {

					/* Burn fuel */
					obj.KGFuel -= obj.TypeP.KgFuelEach

					if obj.MinerData.NumTypesFound > 0 {
						pick := rand.Intn(int(obj.MinerData.NumTypesFound))

						amount := obj.TypeP.KgMineEach * obj.MinerData.MatsFound[pick]
						kind := MatTypes[obj.MinerData.MatsFoundT[pick]]

						/* If we are mining coal, and it won't overfill us,
						 * burn the coal and don't output */
						if obj.MinerData.MatsFoundT[pick] == gv.MAT_COAL &&
							obj.KGFuel+amount <= obj.TypeP.MaxFuelKG {
							obj.KGFuel += amount
						} else {

							obj.Ports[obj.Dir].Buf.Amount = amount
							obj.Ports[obj.Dir].Buf.TypeP = kind
							obj.Ports[obj.Dir].Buf.Rot = uint8(rand.Intn(3))
						}
						//We should remove ourselves here if we run out of ore
					}
				}
				obj.TickCount = 0
			}
		}

	}
}

func beltUpdate(obj *glob.ObjData) {

	/* Output is full, exit */
	if obj.Ports[obj.Dir].Buf.Amount != 0 {
		obj.Blocked = true
		obj.Active = false
		return
	}
	obj.Blocked = false

	/* Find all inputs, round-robin send to output */
	dir := obj.LastUsedInput
	for x := 0; x < 4; x++ {
		dir = util.RotCW(dir)

		if obj.Ports[dir].PortDir != gv.PORT_INPUT {
			continue
		}
		if obj.Ports[dir].Buf.Amount == 0 {
			obj.Active = false
			continue
		} else {
			obj.Active = true
			obj.Ports[obj.Dir].Buf.Amount = obj.Ports[dir].Buf.Amount
			obj.Ports[obj.Dir].Buf.TypeP = obj.Ports[dir].Buf.TypeP
			obj.Ports[obj.Dir].Buf.Rot = obj.Ports[dir].Buf.Rot
			obj.Ports[dir].Buf.Amount = 0
			obj.LastUsedInput = dir
			break
		}

	}
}

func fuelHopperUpdate(obj *glob.ObjData) {

	/* Handle putting fuel into objects */
	if obj.Ports[obj.Dir].Obj != nil &&
		obj.Ports[obj.Dir].Obj.TypeP.MaxFuelKG > 0 {

	} else {
		obj.Blocked = true
	}
}

func splitterUpdate(obj *glob.ObjData) {

	input := util.ReverseDirection(obj.Dir)
	if obj.Ports[input].Buf.Amount == 0 {
		cwlog.DoLog("beltUpdate: Our input is empty. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
		return
	}

	/* Find all inputs, round-robin send to output */
	dir := obj.LastUsedOutput
	for x := 0; x < 4; x++ {
		dir = util.RotCW(dir)

		if obj.Ports[dir].PortDir != gv.PORT_OUTPUT {
			continue
		}
		if obj.Ports[dir].Buf.Amount != 0 {
			continue
		} else {
			obj.Ports[dir].Buf.Amount = obj.Ports[input].Buf.Amount
			obj.Ports[dir].Buf.TypeP = obj.Ports[input].Buf.TypeP
			obj.Ports[dir].Buf.Rot = obj.Ports[input].Buf.Rot
			obj.Ports[input].Buf.Amount = 0
			obj.LastUsedOutput = dir
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

		if obj.KGHeld+port.Buf.Amount > obj.TypeP.MaxContainKG {
			cwlog.DoLog("boxUpdate: Object is full %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			obj.Blocked = true
			continue
		}
		obj.Blocked = false

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
		if port == nil {
			continue
		}
		if port.PortDir == gv.PORT_INPUT {
			if port.Buf.TypeP == nil {
				continue
			}

			/* If this is fuel or ore, take it */
			if port.Buf.TypeP.TypeI == gv.MAT_COAL {
				if obj.KGFuel+port.Buf.Amount > obj.TypeP.MaxFuelKG {
					continue
				}
				obj.KGFuel += port.Buf.Amount
				obj.Ports[p].Buf.Amount = 0
			} else if port.Buf.TypeP.IsOre {
				if obj.KGHeld+port.Buf.Amount > obj.TypeP.MaxContainKG {
					continue
				}

				if obj.Contents[port.Buf.TypeP.TypeI] == nil {
					obj.Contents[port.Buf.TypeP.TypeI] = &glob.MatData{}
					obj.Contents[port.Buf.TypeP.TypeI].TypeP = port.Buf.TypeP
				}

				obj.Contents[port.Buf.TypeP.TypeI].Amount += port.Buf.Amount

				obj.KGHeld += port.Buf.Amount
				obj.Ports[p].Buf.Amount = 0
			}
		} else {

			/* Output is full, exit */
			if port.Buf.Amount != 0 {
				obj.Blocked = true
				//cwlog.DoLog("smelterUpdate: Our output is blocked. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
				continue
			}
			obj.Blocked = false

			obj.TickCount++
			if obj.TickCount >= obj.TypeP.Interval {

				/* Smelt stuff */
				if obj.KGFuel >= obj.TypeP.KgFuelEach {
					for c, cont := range obj.Contents {
						if cont == nil {
							continue
						}

						if cont.TypeP.IsOre && cont.Amount >= obj.TypeP.KgMineEach {
							obj.KGFuel -= obj.TypeP.KgFuelEach

							obj.Contents[c].Amount -= obj.TypeP.KgMineEach
							obj.KGHeld -= obj.TypeP.KgMineEach

							obj.Ports[p].Buf.Amount = obj.TypeP.KgMineEach * gv.ORE_WASTE
							obj.Ports[p].Buf.TypeP = MatTypes[cont.TypeP.Result]
							obj.Ports[p].Buf.Rot = port.Buf.Rot
						}
					}
				}
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
