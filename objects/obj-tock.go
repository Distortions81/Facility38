package objects

import (
	"GameTest/gv"
	"GameTest/world"
	"math/rand"
)

func minerUpdate(obj *world.ObjData) {

	/* Nothing to do, sleep */
	if obj.NumOut == 0 {
		tocklistRemove(obj)
		ticklistRemove(obj)
		return
	}

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
		}
	}

	if obj.KGFuel < obj.TypeP.KgFuelEach {
		/* Not enough fuel, exit */
		if obj.Active {
			obj.Active = false
		}
		return
	}

	for p := range obj.Outputs {

		/* Cycle through available materials */
		var pick uint8 = 0
		if obj.MinerData.ResourcesCount > 1 {
			if obj.MinerData.LastUsed < (obj.MinerData.ResourcesCount - 1) {
				obj.MinerData.LastUsed++
				pick = obj.MinerData.LastUsed
			}
		}

		/* Calculate how much material */
		if obj.MinerData.ResourcesCount == 0 {
			return
		}
		amount := obj.TypeP.KgMineEach * float32(obj.MinerData.Resources[pick])
		kind := MatTypes[obj.MinerData.ResourcesType[pick]]

		/* Stop if the amount is extremely small, zero or negative */
		if amount < 0.001 {
			break
		}

		/* Set as actively working */
		if !obj.Active {
			obj.Active = true
		}

		/* Tally the amount taken as well as the type */
		obj.Tile.MinerData.Mined[pick] += amount

		/* Output the material */
		obj.Outputs[p].Buf.Amount = amount
		if obj.Outputs[p].Buf.TypeP != kind {
			obj.Outputs[p].Buf.TypeP = kind
		}
		obj.Outputs[p].Buf.Rot = uint8(rand.Intn(3))

		/* Burn fuel */
		obj.KGFuel -= obj.TypeP.KgFuelEach

		// TODO: Remove our events if there is nothing left to mine
		break
	}
}

func beltUpdateInter(obj *world.ObjData) {
}

func beltUpdate(obj *world.ObjData) {

	/* Nothing to do, sleep */
	if obj.NumOut == 0 || obj.NumIn == 0 {
		tocklistRemove(obj)
		ticklistRemove(obj)
		return
	} else if obj.NumIn > 1 {
		if obj.LastInput == (obj.NumIn - 1) {
			obj.LastInput = 0
		} else {
			obj.LastInput++
		}
	}

	/* Does the input contain anything? */
	if obj.Inputs[obj.LastInput].Buf.Amount > 0 {
		/* Good to go, swap pointers */
		*obj.Outputs[0].Buf, *obj.Inputs[obj.LastInput].Buf = *obj.Inputs[obj.LastInput].Buf, *obj.Outputs[0].Buf
	}
}

func fuelHopperUpdate(obj *world.ObjData) {

	/* Nothing to do, sleep */
	if obj.NumFOut == 0 || obj.NumIn == 0 {
		tocklistRemove(obj)
		ticklistRemove(obj)
		return
	}

	/* Is it time to run? */
	if obj.TickCount < obj.TypeP.Interval {
		/* Increment timer */
		obj.TickCount++
		return
	}
	obj.TickCount = 0

	for i, input := range obj.Inputs {

		/* Does input contain anything? */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Is input solid? */
		if !input.Buf.TypeP.IsSolid {
			continue
		}

		/* Is input fuel? */
		if !input.Buf.TypeP.IsFuel {
			continue
		}

		/* Do we have room for it? */
		if (obj.KGFuel + input.Buf.Amount) < obj.TypeP.MaxFuelKG {
			obj.KGFuel += input.Buf.Amount
			obj.Inputs[i].Buf.Amount = 0
			break
		}
	}

	/* Grab destination object */
	if obj.KGFuel > (obj.TypeP.KgHopperMove + obj.TypeP.KgFuelEach) {
		for _, output := range obj.FuelOut {
			output.Buf.Amount = obj.TypeP.KgHopperMove
			obj.KGFuel -= (obj.TypeP.KgHopperMove + obj.TypeP.KgFuelEach)
			break
		}
	}

}

func splitterUpdate(obj *world.ObjData) {

	/* Nothing to do, sleep */
	if obj.NumIn == 0 || obj.NumOut == 0 {
		tocklistRemove(obj)
		ticklistRemove(obj)
		return
	}
	if obj.Outputs[0].Buf.Amount != 0 {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	var x uint8
	if obj.LastInput == obj.NumIn {
		x = 0
	}
	for x = 0; x < 3; x++ {
		xp := obj.LastInput % 3
		if obj.Inputs[xp].Buf.Amount != 0 {
			/* Swap pointers */
			*obj.Outputs[0].Buf, *obj.Inputs[xp].Buf = *obj.Inputs[xp].Buf, *obj.Outputs[0].Buf
			obj.LastInput = xp

			if !obj.Active {
				obj.Active = true
			}
			return
		}
	}

	if obj.Active {
		obj.Active = false
	}
}

func boxUpdate(obj *world.ObjData) {

	/* Nothing to do, sleep */
	if obj.NumIn == 0 {
		tocklistRemove(obj)
		ticklistRemove(obj)
		return
	}
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
			obj.Active = false
			continue
		}

		/* Reset counter */
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

	/* Nothing to do, sleep */
	if obj.NumIn == 0 || obj.NumOut == 0 {
		tocklistRemove(obj)
		ticklistRemove(obj)
		return
	}

	/* Output full? */
	for _, output := range obj.Outputs {
		if output.Buf.Amount != 0 {
			return
		}
	}

	/* Check input */
	for i, input := range obj.Inputs {

		/* Input contains something */
		if input.Buf.Amount != 0 {
			/* Is of a type we can smelt */
			if input.Buf.TypeP.IsSolid {
				/* If the material will fit */
				if input.Buf.Amount <= obj.TypeP.MaxContainKG {
					/* If type is nil, init */
					if obj.SContent.TypeP == nil {
						obj.SContent.TypeP = input.Buf.TypeP
						/* If we already contain ore */
					} else if obj.SContent.Amount > 0 {
						/* If of different type, ruin the contents */
						if input.Buf.TypeP.TypeI != obj.SContent.TypeP.TypeI {
							obj.SContent.TypeP = MatTypes[gv.MAT_MIXORE]
						}
					} else {
						/* Otherwise, just set the ore type */
						obj.SContent.TypeP = input.Buf.TypeP
					}

					/* Add contents */
					obj.SContent.Amount += input.Buf.Amount

					/* Add to content weight */
					obj.KGHeld += input.Buf.Amount

					/* Clear input */
					obj.Inputs[i].Buf.Amount = 0
				}
			}
		}
	}

}

func ironCasterUpdate(o *world.ObjData) {
}

func steamEngineUpdate(o *world.ObjData) {
}
