package objects

import (
	"GameTest/gv"
	"GameTest/world"
	"math/rand"
)

func minerUpdate(obj *world.ObjData) {

	/* Time to run? */
	if obj.TickCount < obj.TypeP.Interval {
		obj.TickCount++
		return
	}
	obj.TickCount = 0

	/* Get fuel */
	for p, port := range obj.FuelIn {
		/* Will it over fill us? */
		if port.Buf.Amount > 0 &&
			obj.KGFuel+port.Buf.Amount <= obj.TypeP.MaxFuelKG {

			/* Eat the fuel */
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

	if obj.NumOut == 0 {
		return
	}

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
		return
	}

	/* Set as actively working */
	if !obj.Active {
		obj.Active = true
	}

	/* Tally the amount taken as well as the type */
	obj.Tile.MinerData.Mined[pick] += amount

	/* Output the material */
	obj.Outputs[0].Buf.Amount = amount
	if obj.Outputs[0].Buf.TypeP != kind {
		obj.Outputs[0].Buf.TypeP = kind
	}
	obj.Outputs[0].Buf.Rot = uint8(rand.Intn(3))

	/* Burn fuel */
	obj.KGFuel -= obj.TypeP.KgFuelEach

}

func beltUpdateOver(obj *world.ObjData) {

	for i, input := range obj.Inputs {
		/* Does the input contain anything? */
		if input.Buf.Amount > 0 &&
			obj.NumOut >= uint8(i+1) &&
			obj.Outputs[i].Buf.Amount == 0 &&
			obj.Outputs[i].Obj != nil &&
			!obj.Outputs[i].Obj.Blocked {
			/* Good to go, swap pointers */
			*obj.Outputs[i].Buf, *obj.Inputs[i].Buf = *obj.Inputs[i].Buf, *obj.Outputs[i].Buf
		}
	}
}

func beltUpdate(obj *world.ObjData) {

	if obj.NumIn > 1 {
		if obj.LastInput == (obj.NumIn - 1) {
			obj.LastInput = 0
		} else {
			obj.LastInput++
		}
	}

	/* Does the input contain anything? */
	if obj.Inputs[obj.LastInput].Buf.Amount > 0 &&
		obj.Outputs[0].Buf.Amount == 0 &&
		obj.Outputs[0].Obj != nil &&
		!obj.Outputs[0].Obj.Blocked {
		/* Good to go, swap pointers */
		*obj.Outputs[0].Buf, *obj.Inputs[obj.LastInput].Buf = *obj.Inputs[obj.LastInput].Buf, *obj.Outputs[0].Buf
	}
}

func fuelHopperUpdate(obj *world.ObjData) {

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

	/* Is it time to run? */
	if obj.TickCount < obj.TypeP.Interval {
		/* Increment timer */
		obj.TickCount++
		return
	}
	obj.TickCount = 0

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

	if obj.Inputs[0].Buf.Amount > 0 {
		if obj.NumOut > 1 {
			if obj.LastOutput >= (obj.NumOut - 1) {
				obj.LastOutput = 0
			} else {
				obj.LastOutput++
			}
		}

		if obj.Outputs[obj.LastOutput].Buf.Amount == 0 {
			/* Good to go, swap pointers */
			*obj.Inputs[0].Buf, *obj.Outputs[obj.LastOutput].Buf = *obj.Outputs[obj.LastOutput].Buf, *obj.Inputs[0].Buf
			if !obj.Active {
				obj.Active = true
			}
			return
		}

		if obj.Active {
			obj.Active = false
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
			obj.Active = false
			continue
		}

		/* Reset counter */
		obj.Active = true
		obj.TickCount = 0

		/* Init content type if needed */
		if obj.Contents.Mats[port.Buf.TypeP.TypeI] == nil {
			obj.Contents.Mats[port.Buf.TypeP.TypeI] = &world.MatData{}
		}

		/* Add to contents */
		obj.Contents.Mats[port.Buf.TypeP.TypeI].Amount += obj.Inputs[p].Buf.Amount
		obj.Contents.Mats[port.Buf.TypeP.TypeI].TypeP = MatTypes[port.Buf.TypeP.TypeI]
		obj.KGHeld += port.Buf.Amount
		obj.Inputs[p].Buf.Amount = 0
		continue

		//Unloader goes here
	}
}

func smelterUpdate(obj *world.ObjData) {

	/* Check input */
	for _, input := range obj.Inputs {

		/* Input contains something */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Contents are solid */
		if !input.Buf.TypeP.IsSolid {
			continue
		}

		/* Process fuel */
		if input.Buf.TypeP.IsFuel {

			/* Will the fuel fit? */
			if obj.KGFuel+input.Buf.Amount > obj.TypeP.MaxFuelKG {
				continue
			}

			obj.KGFuel += input.Buf.Amount
			input.Buf.Amount = 0
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+input.Buf.Amount > obj.TypeP.MaxContainKG {
			continue
		}

		/* Are we mixing materials? */
		if obj.SingleContent.Amount > 0 &&
			input.Buf.TypeP != obj.SingleContent.TypeP {
			obj.SingleContent.TypeP = MatTypes[gv.MAT_MIXORE]

			/* If not, set material type if needed */
		} else if obj.SingleContent.TypeP != input.Buf.TypeP {
			obj.SingleContent.TypeP = input.Buf.TypeP
		}

		/* Add to weight */
		obj.KGHeld += input.Buf.Amount

		/* Add input to contents */
		obj.SingleContent.Amount += input.Buf.Amount
		input.Buf.Amount = 0
	}

	/* Is it time to run? */
	if obj.TickCount < obj.TypeP.Interval {
		/* Increment timer */
		obj.TickCount++
		return
	}
	obj.TickCount = 0

	/* Process ores */

	/* Do we have valid ore material? */
	if obj.SingleContent.TypeP.TypeI == gv.MAT_MIXORE {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Is there enough ore to process? */
	if obj.SingleContent.Amount < obj.TypeP.KgMineEach {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.KGFuel < obj.TypeP.KgFuelEach {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	if !obj.Active {
		obj.Active = true
	}

	/* Burn fuel */
	obj.KGFuel -= obj.TypeP.KgFuelEach

	/* Subtract ore */
	obj.SingleContent.Amount -= obj.TypeP.KgMineEach
	/* Subtract ore weight */
	obj.KGHeld -= obj.TypeP.KgMineEach

	/* Output result */
	obj.Outputs[0].Buf.Amount = obj.TypeP.KgMineEach

	/* Find and set result type, if needed */
	result := MatTypes[obj.SingleContent.TypeP.Result]
	if obj.Outputs[0].Buf.TypeP != result {
		obj.Outputs[0].Buf.TypeP = result
	}
}
