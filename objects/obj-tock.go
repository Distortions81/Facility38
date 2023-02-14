package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"math"
	"math/rand"
)

func toggleOverlay() {
	if world.ShowInfoLayer {
		world.ShowInfoLayer = false
	} else {
		world.ShowInfoLayer = true
	}
}

func InitMiner(obj *world.ObjData) {
	if obj == nil {
		return
	}

	/* Init miner data if needed */
	if obj.MinerData == nil {
		obj.MinerData = &world.MinerDataType{}
	} else {
		return
	}

	/* Check for resources to mine */
	for p := 1; p < len(NoiseLayers); p++ {
		var h float32 = float32(math.Abs(float64(NoiseMap(float32(obj.Pos.X), float32(obj.Pos.Y), p))))

		/* We only mind solids */
		if !NoiseLayers[p].TypeP.IsSolid {
			continue
		}
		if h > 0 {
			obj.MinerData.MatsFound = append(obj.MinerData.MatsFound, h)
			obj.MinerData.MatsFoundT = append(obj.MinerData.MatsFoundT, NoiseLayers[p].TypeI)
			obj.MinerData.NumMatsFound++
		}
	}
}

func minerUpdate(obj *world.ObjData) {

	/* Valid? */
	if obj.MinerData == nil {
		obj.Blocked = true
		return
	}

	/* Anything to mine? */
	if obj.MinerData.NumMatsFound == 0 {
		obj.Blocked = true
		return
	}

	/* Cycle all ports */
	for p, port := range obj.Ports {
		/* Valid? */
		if port == nil {
			continue
		}

		/* Fuel input */
		if port.PortDir == gv.PORT_INPUT {

			/* Valid? */
			if port.Buf.TypeP == nil {
				continue
			}

			/* Is it fuel? */
			if port.Buf.TypeP.TypeI != gv.MAT_COAL {
				continue
			}

			/* Will it over fill us? */
			if obj.KGFuel+port.Buf.Amount > obj.TypeP.MaxFuelKG {
				continue
			}

			/* Eat the fuel and increase fuel kg */
			obj.KGFuel += port.Buf.Amount
			obj.Ports[p].Buf.Amount = 0
			continue
		}

		/* Output full? */
		if port.Buf.Amount != 0 {
			obj.Blocked = true
			obj.Active = false
			continue
		}

		/* Then we are not blocked */
		obj.Blocked = false
		/* Increment timer */
		obj.TickCount++
		/* Turn on active status */
		obj.Active = true

		/* Are we ready to output yet? */
		if obj.TickCount < obj.TypeP.Interval {
			continue
		}

		/* Randomly pick a material from the list */
		pick := rand.Intn(int(obj.MinerData.NumMatsFound))

		/* Calculate how much material */
		amount := obj.TypeP.KgMineEach * float32(obj.MinerData.MatsFound[pick])
		kind := MatTypes[obj.MinerData.MatsFoundT[pick]]

		/* Are we are mining coal? */
		if obj.MinerData.MatsFoundT[pick] == gv.MAT_COAL &&
			obj.KGFuel+amount <= obj.TypeP.MaxFuelKG {

			/* If we need fuel, fuel ourselves */
			obj.KGFuel += amount
		} else {
			if obj.KGFuel < obj.TypeP.KgFuelEach {
				/* Not enough fuel */
				continue
			}
			/* Otherwise output the material */
			obj.Ports[obj.Dir].Buf.Amount = amount
			obj.Ports[obj.Dir].Buf.TypeP = kind
			obj.Ports[obj.Dir].Buf.Rot = uint8(rand.Intn(3))
		}

		/* Burn fuel */
		obj.KGFuel -= obj.TypeP.KgFuelEach

		//We should remove ourselves here if we run out of ore

		obj.TickCount = 0

	}
}

func beltUpdateInter(obj *world.ObjData) {

	for p, port := range obj.Ports {
		if port == nil {
			continue
		}
		if port.PortDir != gv.PORT_INPUT {
			continue
		}
		if port.Obj == nil {
			continue
		}
		odir := util.ReverseDirection(uint8(p))
		if obj.Ports[odir] == nil {
			continue
		}
		if obj.Ports[odir].Obj == nil {
			continue
		}
		if obj.Ports[odir].PortDir != gv.PORT_OUTPUT {
			continue
		}
		if obj.Ports[p].Buf.Amount > 0 && obj.Ports[odir].Buf.Amount == 0 {
			obj.Ports[odir].Buf.Amount = obj.Ports[p].Buf.Amount
			obj.Ports[odir].Buf.TypeP = obj.Ports[p].Buf.TypeP
			obj.Ports[odir].Buf.Rot = obj.Ports[p].Buf.Rot

			obj.Ports[p].Buf.Amount = 0
		}
	}

}

func beltUpdate(obj *world.ObjData) {

	/* Output full? */
	if obj.Ports[obj.Dir].Buf.Amount != 0 {
		obj.Blocked = true
		obj.Active = false
		return
	}
	obj.Blocked = false

	/* Find all inputs round-robin, send to output */
	dir := obj.LastUsedInput
	/* Start with last input, then rotate one */
	for x := 0; x < 4; x++ {
		dir = util.RotCW(dir)

		/* Is this an input? */
		if obj.Ports[dir].PortDir != gv.PORT_INPUT {
			continue
		}

		/* Does the input contain anything? */
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
			break /* Stop */
		}

	}
}

func fuelHopperUpdate(obj *world.ObjData) {

	/* Valid port? */
	if obj.Ports[obj.Dir] == nil {
		obj.Blocked = true
		return
	}

	/* Connected to valid object? */
	if obj.Ports[obj.Dir].Obj == nil {
		obj.Blocked = true
		return
	}

	/* Grab destination object */
	dest := obj.Ports[obj.Dir].Obj

	/* Does it use fuel? */
	if dest.TypeP.MaxFuelKG == 0 {
		obj.Blocked = true
		return
	}
}

func splitterUpdate(obj *world.ObjData) {

	input := util.ReverseDirection(obj.Dir)

	/* Anything in the input? */
	if obj.Ports[input].Buf.Amount == 0 {
		obj.Active = false
		return
	}

	/* Round-robin output */
	dir := obj.LastUsedOutput
	for x := 0; x < 4; x++ {
		dir = util.RotCW(dir)

		if obj.Ports[dir].PortDir != gv.PORT_OUTPUT {
			continue
		}
		if obj.Ports[dir].Buf.Amount != 0 {
			obj.Active = false
			continue
		} else {
			obj.Ports[dir].Buf.Amount = obj.Ports[input].Buf.Amount
			obj.Ports[dir].Buf.TypeP = obj.Ports[input].Buf.TypeP
			obj.Ports[dir].Buf.Rot = obj.Ports[input].Buf.Rot
			obj.Ports[input].Buf.Amount = 0
			obj.LastUsedOutput = dir
			obj.Active = true
			break
		}

	}

}

func boxUpdate(obj *world.ObjData) {
	for p, port := range obj.Ports {
		if port.PortDir == gv.PORT_INPUT {

			if port.Buf.Amount == 0 {
				//cwlog.DoLog("tickObj: Input is empty. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
				if obj.TickCount > uint8(world.ObjectUPS*2) {
					obj.Active = false
				}
				obj.TickCount++
				continue
			}

			if obj.KGHeld+port.Buf.Amount > obj.TypeP.MaxContainKG {
				//cwlog.DoLog("boxUpdate: Object is full %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
				obj.Blocked = true
				obj.Active = false
				continue
			}
			obj.Blocked = false
			obj.Active = true
			obj.TickCount = 0

			if obj.Contents[port.Buf.TypeP.TypeI] == nil {
				obj.Contents[port.Buf.TypeP.TypeI] = &world.MatData{}
			}
			obj.Contents[port.Buf.TypeP.TypeI].Amount += obj.Ports[p].Buf.Amount
			obj.Contents[port.Buf.TypeP.TypeI].TypeP = obj.Ports[p].Buf.TypeP
			obj.KGHeld += port.Buf.Amount
			obj.Ports[p].Buf.Amount = 0

		}

		//Unloader goes here
	}
}

func smelterUpdate(obj *world.ObjData) {

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
			} else if port.Buf.TypeP.IsSolid {
				if obj.KGHeld+port.Buf.Amount > obj.TypeP.MaxContainKG {
					continue
				}

				if obj.Contents[port.Buf.TypeP.TypeI] == nil {
					obj.Contents[port.Buf.TypeP.TypeI] = &world.MatData{}
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

			/* Smelt stuff */
			if obj.KGFuel >= obj.TypeP.KgFuelEach {
				for c, cont := range obj.Contents {
					if cont == nil {
						continue
					}

					if cont.TypeP.IsSolid && cont.Amount >= obj.TypeP.KgMineEach {
						obj.Active = true
						obj.TickCount++
						if obj.TickCount >= obj.TypeP.Interval {
							obj.KGFuel -= obj.TypeP.KgFuelEach

							obj.Contents[c].Amount -= obj.TypeP.KgMineEach
							obj.KGHeld -= obj.TypeP.KgMineEach

							obj.Ports[p].Buf.Amount = obj.TypeP.KgMineEach * gv.ORE_WASTE
							obj.Ports[p].Buf.TypeP = MatTypes[cont.TypeP.Result]
							obj.Ports[p].Buf.Rot = port.Buf.Rot
							obj.TickCount = 0
							obj.Active = false
						}
					}
				}

			}
		}

	}
}

func ironCasterUpdate(o *world.ObjData) {
	//oData := world.GameObjTypes[Obj.Type]

}

func steamEngineUpdate(o *world.ObjData) {
}
