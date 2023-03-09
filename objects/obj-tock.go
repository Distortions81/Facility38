package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"math/rand"
)

func minerUpdate(obj *world.ObjData) {

	/* Is it time to run? */
	if obj.TickCount < obj.TypeP.Interval {
		/* Increment timer */
		obj.TickCount++
		return
	}
	obj.TickCount = 0

	/* Get fuel */
	for p, port := range obj.FuelIn {
		/* Will it over fill us? */
		if port.Buf.Amount > 0 &&
			obj.KGFuel+port.Buf.Amount <= obj.TypeP.MaxFuelKG {

			/* Eat the fuel and increase fuel kg */
			obj.KGFuel += port.Buf.Amount
			obj.FuelIn[p].Buf.Amount = 0
			break
		}
	}

	for p, port := range obj.Outputs {

		/* Output full? */
		if port.Buf.Amount != 0 {
			obj.Blocked = true
			obj.Active = false
			continue
		}

		/* We are not blocked */
		obj.Blocked = false
		obj.Active = true

		if obj.KGFuel < obj.TypeP.KgFuelEach {
			/* Not enough fuel */
			break
		}

		/* Cycle through available materials */
		var pick uint8 = 0
		if obj.MinerData.ResourcesCount > 1 {
			if obj.MinerData.LastUsed < (obj.MinerData.ResourcesCount - 1) {
				obj.MinerData.LastUsed++
			} else {
				obj.MinerData.LastUsed = 0
			}
			pick = obj.MinerData.LastUsed
		} else if obj.MinerData.ResourcesCount == 0 {
			return
		}

		/* Calculate how much material */
		amount := obj.TypeP.KgMineEach * float32(obj.MinerData.Resources[pick])
		kind := MatTypes[obj.MinerData.ResourcesType[pick]]

		/* Stop if the amount is extremely small, zero or negative */
		if amount <= 0.001 {
			break
		}

		/* Tally the amount taken as well as the type */
		obj.Tile.MinerData.Mined[pick] += amount

		/* Otherwise output the material */
		if obj.Outputs[p].Buf.Amount != amount {
			obj.Outputs[p].Buf.Amount = amount
		}
		if obj.Outputs[p].Buf.TypeP != kind {
			obj.Outputs[p].Buf.TypeP = kind
		}
		obj.Outputs[p].Buf.Rot = uint8(rand.Intn(3))

		/* Burn fuel */
		obj.KGFuel -= obj.TypeP.KgFuelEach

		//We should remove ourselves here if we run out of ore
		break
	}
}

func beltUpdateInter(obj *world.ObjData) {

	for _, port := range obj.Inputs {

		/* Revese direction */
		//odir := util.ReverseDirection(uint8(p))

		/* Do we have input and is output is empty */
		if port.Buf.Amount > 0 && port.Buf.Amount == 0 {
		}
	}

}

func beltUpdate(obj *world.ObjData) {

	if obj.NumOut == 0 || obj.NumIn == 0 {
		return
	}

	/* Output full? */
	for _, output := range obj.Outputs {
		if output.Buf.Amount != 0 {
			obj.Blocked = true
			return
		}
	}

	obj.Blocked = false

	/* Loop ports */
	if obj.LastInput == obj.NumIn {
		obj.LastInput = 0
	}
	for x := obj.LastInput; x < obj.NumIn; x++ {

		/* Does the input contain anything? */
		if obj.Inputs[x].Buf.Amount == 0 {
			continue
			/* If the destination blocked, stop */
		} else if obj.Outputs[0].Obj != nil && obj.Outputs[0].Obj.Blocked {
			break
		} else {
			/* Otherwise, do the thing. */
			swapPortBuf(obj.Outputs[0].Buf, obj.Inputs[x].Buf)
			obj.Inputs[x].Buf.Amount = 0
			obj.LastInput = x
			break /* Stop */
		}
	}
}

func fuelHopperUpdate(obj *world.ObjData) {

	/* Grab destination object */
	for _, output := range obj.Outputs {

		if output.Obj == nil {
			continue
		}
		/* Does it use fuel? */
		if output.Obj.TypeP.MaxFuelKG == 0 {
			obj.Blocked = true
			return
		}

	}
}

func splitterUpdate(obj *world.ObjData) {

	/* TODO: Update me for new port system */
	input := util.ReverseDirection(obj.Dir)

	/* Anything in the input? */
	if obj.Ports[input].Buf.Amount == 0 {
		obj.Active = false
		return
	}

	/* Round-robin output */
	dir := obj.LastInput
	for x := 0; x < 4; x++ {
		dir = util.RotCW(dir)

		/* Is this a output? */
		if obj.Ports[dir].Dir != gv.PORT_OUT {
			continue
		}

		/* Is the port empty? */
		if obj.Ports[dir].Buf.Amount != 0 {
			obj.Active = false
			continue
		} else {
			/* Do the thing */
			swapPortBuf(obj.Ports[dir].Buf, obj.Ports[input].Buf)

			obj.Ports[input].Buf.Amount = 0
			obj.LastInput = dir
			obj.Active = true
			break
			/* End */
		}

	}

}

func boxUpdate(obj *world.ObjData) {
	for p, port := range obj.Inputs {
		if port.Buf.Amount == 0 {

			/* Go inactive after a while */
			if obj.TickCount > uint8(world.ObjectUPS*4) {
				obj.Active = false
			}
			obj.TickCount++
			continue
		}

		/* Will the input fit? */
		if obj.KGHeld+port.Buf.Amount > obj.TypeP.MaxContainKG {
			obj.Blocked = true
			obj.Active = false
			continue
		}

		/* Reset counter */
		obj.Blocked = false
		obj.Active = true
		obj.TickCount = 0

		/* Init content type if needed */
		if obj.Contents[port.Buf.TypeP.TypeI] == nil {
			obj.Contents[port.Buf.TypeP.TypeI] = &world.MatData{}
		}

		/* Add to contents */
		obj.Contents[port.Buf.TypeP.TypeI].Amount += obj.Inputs[p].Buf.Amount
		obj.Contents[port.Buf.TypeP.TypeI].TypeP = MatTypes[port.Buf.TypeP.TypeI]
		obj.KGHeld += port.Buf.Amount
		obj.Inputs[p].Buf.Amount = 0
		continue

		//Unloader goes here
	}
}

func smelterUpdate(obj *world.ObjData) {

}

func oldSmelterUpdate(obj *world.ObjData) {

	for p, port := range obj.Ports {

		/* Input? */
		if port.Dir == gv.PORT_IN {

			/* Valid input? */
			if port.Buf.TypeP == nil {
				continue
			}

			/* Is this fuel? */
			if port.Buf.TypeP.TypeI == gv.MAT_COAL {

				/* Will it fit? */
				if obj.KGFuel+port.Buf.Amount <= obj.TypeP.MaxFuelKG {
					obj.KGFuel += port.Buf.Amount
					obj.Ports[p].Buf.Amount = 0
				}

				/* Is this ore? */
			} else if port.Buf.TypeP.IsSolid {
				if obj.KGHeld+port.Buf.Amount > obj.TypeP.MaxContainKG {
					continue
				}

				/* Init content type if needed */
				if obj.Contents[port.Buf.TypeP.TypeI] == nil {
					obj.Contents[port.Buf.TypeP.TypeI] = &world.MatData{}
					obj.Contents[port.Buf.TypeP.TypeI].TypeP = port.Buf.TypeP
				}

				/* Add contents */
				obj.Contents[port.Buf.TypeP.TypeI].Amount += port.Buf.Amount

				/* Add to content weight */
				obj.KGHeld += port.Buf.Amount

				/* Clear input */
				obj.Ports[p].Buf.Amount = 0
			}
		} else {

			/* Is the output empty? */
			if port.Buf.Amount != 0 {
				obj.Blocked = true
				continue
			}
			obj.Blocked = false

			/* Do we have enough fuel to mine? */
			if obj.KGFuel >= obj.TypeP.KgFuelEach {
				for c, cont := range obj.Contents {

					/* Valid contents? */
					if cont == nil {
						continue
					}

					/* Is it ore, and do we have enough to process? */
					if cont.TypeP.IsSolid && cont.Amount >= obj.TypeP.KgMineEach {
						obj.Active = true
						obj.TickCount++

						/* Are we finished? */
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
}

func steamEngineUpdate(o *world.ObjData) {
}
