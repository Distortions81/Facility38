package main

import (
	"Facility38/cwlog"
	"Facility38/gv"
	"Facility38/world"
	"math/rand"
)

func minerUpdate(obj *world.ObjData) {

	/* Get fuel */
	for p, port := range obj.FuelIn {
		/* Will it over fill us? */
		if port.Buf.Amount > 0 &&
			obj.Unique.KGFuel+port.Buf.Amount <= obj.Unique.TypeP.MachineSettings.MaxFuelKG {

			/* Eat the fuel */
			obj.Unique.KGFuel += port.Buf.Amount
			obj.FuelIn[p].Buf.Amount = 0
		}
	}

	if obj.Unique.KGFuel < obj.Unique.TypeP.MachineSettings.KgFuelPerCycle {
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
	amount := obj.Unique.TypeP.MachineSettings.KgPerCycle * float32(obj.MinerData.Resources[pick])
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
	obj.Unique.KGFuel -= obj.Unique.TypeP.MachineSettings.KgFuelPerCycle

}

func beltUpdateOver(obj *world.ObjData) {

	/* Underpass */
	if obj.BeltOver.UnderIn != nil && obj.BeltOver.UnderOut != nil {
		if obj.BeltOver.UnderOut.Obj != nil && !obj.BeltOver.UnderOut.Obj.Blocked {
			if obj.BeltOver.UnderIn.Buf.Amount != 0 && obj.BeltOver.UnderOut.Buf.Amount == 0 {
				*obj.BeltOver.UnderOut.Buf, *obj.BeltOver.UnderIn.Buf = *obj.BeltOver.UnderIn.Buf, *obj.BeltOver.UnderOut.Buf
			}
		}
	}

	/* Overpass to OverOut */
	if obj.BeltOver.OverOut != nil && obj.BeltOver.Middle != nil {
		if obj.BeltOver.OverOut.Obj != nil {
			if obj.BeltOver.Middle.Amount != 0 && obj.BeltOver.OverOut.Buf.Amount == 0 {
				*obj.BeltOver.OverOut.Buf, *obj.BeltOver.Middle = *obj.BeltOver.Middle, *obj.BeltOver.OverOut.Buf
			}
		}
	}

	/* OverIn to Overpass */
	if obj.BeltOver.OverIn != nil && obj.BeltOver.Middle != nil {
		if obj.BeltOver.OverIn.Buf.Amount != 0 && obj.BeltOver.Middle.Amount == 0 {
			*obj.BeltOver.Middle, *obj.BeltOver.OverIn.Buf = *obj.BeltOver.OverIn.Buf, *obj.BeltOver.Middle
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
	if obj.NumOut > 0 && obj.NumIn > 0 {
		if obj.Inputs[obj.LastInput].Buf.Amount > 0 &&
			obj.Outputs[0].Buf.Amount == 0 &&
			obj.Outputs[0].Obj != nil &&
			!obj.Outputs[0].Obj.Blocked {
			/* Good to go, swap pointers */
			*obj.Outputs[0].Buf, *obj.Inputs[obj.LastInput].Buf = *obj.Inputs[obj.LastInput].Buf, *obj.Outputs[0].Buf
		}
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
		if (obj.Unique.KGFuel + input.Buf.Amount) < obj.Unique.TypeP.MachineSettings.MaxFuelKG {
			obj.Unique.KGFuel += input.Buf.Amount
			obj.Inputs[i].Buf.Amount = 0
			break
		}
	}

	if obj.Unique.KGFuel > 0 {
		if !obj.Active {
			obj.Active = true
		}
	} else {
		if obj.Active {
			obj.Active = false
		}
	}

	/* Grab destination object */
	if obj.Unique.KGFuel > (obj.Unique.TypeP.MachineSettings.KgHopperMove + obj.Unique.TypeP.MachineSettings.KgFuelPerCycle) {
		for _, output := range obj.FuelOut {
			output.Buf.Amount = obj.Unique.TypeP.MachineSettings.KgHopperMove
			obj.Unique.KGFuel -= (obj.Unique.TypeP.MachineSettings.KgHopperMove + obj.Unique.TypeP.MachineSettings.KgFuelPerCycle)
			break
		}
	}
}

func loaderUpdate(obj *world.ObjData) {
	for i, input := range obj.Inputs {
		if input.Buf.Amount == 0 {
			continue
		}
		if obj.NumOut == 0 || obj.Outputs[0].Buf.Amount != 0 {
			continue
		}
		*obj.Outputs[0].Buf, *obj.Inputs[i].Buf = *obj.Inputs[i].Buf, *obj.Outputs[0].Buf
		break
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
		} else {
			return
		}

		if obj.Outputs[obj.LastOutput].Buf.Amount == 0 {
			/* Good to go, swap pointers */
			*obj.Inputs[0].Buf, *obj.Outputs[obj.LastOutput].Buf = *obj.Outputs[obj.LastOutput].Buf, *obj.Inputs[0].Buf
			return
		}
	}
}

func boxUpdate(obj *world.ObjData) {

	for p, port := range obj.Inputs {
		if port.Buf.TypeP == nil {
			continue
		}

		/* Will the input fit? */
		if obj.KGHeld+port.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxContainKG {
			obj.Active = false
			continue
		}

		/* Init content type if needed */
		if obj.Unique.Contents.Mats[port.Buf.TypeP.TypeI] == nil {
			obj.Unique.Contents.Mats[port.Buf.TypeP.TypeI] = &world.MatData{}
		}

		/* Add to contents */
		obj.Unique.Contents.Mats[port.Buf.TypeP.TypeI].Amount += obj.Inputs[p].Buf.Amount
		obj.Unique.Contents.Mats[port.Buf.TypeP.TypeI].TypeP = MatTypes[port.Buf.TypeP.TypeI]
		obj.KGHeld += port.Buf.Amount
		obj.Inputs[p].Buf.Amount = 0
		continue

	}
}

func smelterUpdate(obj *world.ObjData) {

	/* Get fuel */
	for _, fuel := range obj.FuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
	}

	/* Check input */
	for _, input := range obj.Inputs {

		/* Input contains something */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Input is ore */
		if !input.Buf.TypeP.IsOre {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+input.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.TypeP != input.Buf.TypeP {
			if obj.Unique.SingleContent.Amount > 0 {
				obj.Unique.SingleContent.TypeP = MatTypes[gv.MAT_MIX_ORE]
			} else {
				obj.Unique.SingleContent.TypeP = input.Buf.TypeP
			}
		}

		/* Add to weight */
		obj.KGHeld += input.Buf.Amount

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += input.Buf.Amount
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < obj.Unique.TypeP.MachineSettings.KgPerCycle {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.TypeP.MachineSettings.KgFuelPerCycle {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	if !obj.Active {
		obj.Active = true
	}

	/* Look up material */
	rec := obj.Unique.TypeP.RecipieLookup[obj.Unique.SingleContent.TypeP.TypeI]
	if rec == nil {
		cwlog.DoLog(true, "Nil recipie")
		return
	}
	result := rec.ResultP[0]

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.TypeP.MachineSettings.KgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount -= obj.Unique.TypeP.MachineSettings.KgPerCycle
	/* Subtract ore weight */
	obj.KGHeld -= obj.Unique.TypeP.MachineSettings.KgPerCycle

	/* Output result */
	obj.Outputs[0].Buf.Amount = obj.Unique.TypeP.MachineSettings.KgPerCycle

	/* Find and set result type, if needed */
	if obj.Outputs[0].Buf.TypeP != result {
		obj.Outputs[0].Buf.TypeP = result
	}
}

func casterUpdate(obj *world.ObjData) {

	/* Get fuel */
	for _, fuel := range obj.FuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
		continue
	}

	/* Check input */
	for _, input := range obj.Inputs {

		/* Input contains something */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Contents are shot */
		if !input.Buf.TypeP.IsShot {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+input.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.TypeP != input.Buf.TypeP {
			obj.Unique.SingleContent.TypeP = input.Buf.TypeP
		}

		/* Add to weight */
		obj.KGHeld += input.Buf.Amount

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += input.Buf.Amount
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < 1 {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Process ores */
	/* Is there enough ore to process? */
	rec := obj.Unique.TypeP.RecipieLookup[obj.Unique.SingleContent.TypeP.TypeI]
	if rec == nil {
		cwlog.DoLog(true, "Nil recipie")
		return
	}
	result := rec.ResultP[0]

	if obj.Unique.SingleContent.Amount < result.KG {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.TypeP.MachineSettings.KgFuelPerCycle {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	if !obj.Active {
		obj.Active = true
	}

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.TypeP.MachineSettings.KgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount -= result.KG
	/* Subtract ore weight */
	obj.KGHeld -= result.KG

	/* Output result */
	obj.Outputs[0].Buf.Amount = result.KG

	/* Find and set result type, if needed */
	if obj.Outputs[0].Buf.TypeP != result {
		obj.Outputs[0].Buf.TypeP = result
	}
}

func rodCasterUpdate(obj *world.ObjData) {

	/* Get fuel */
	for _, fuel := range obj.FuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
		continue
	}

	/* Check input */
	for _, input := range obj.Inputs {

		/* Input contains something */
		if input.Buf.Amount < 1 {
			continue
		}

		/* Contents is metal bar */
		if !input.Buf.TypeP.IsBar {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+(input.Buf.Amount) > obj.Unique.TypeP.MachineSettings.MaxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.TypeP != input.Buf.TypeP {
			obj.Unique.SingleContent.TypeP = input.Buf.TypeP
		}

		/* Add to weight */
		obj.KGHeld += (input.Buf.Amount * input.Buf.TypeP.KG)

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += (input.Buf.Amount * input.Buf.TypeP.KG)
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < 1 {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.TypeP.MachineSettings.KgFuelPerCycle {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	if !obj.Active {
		obj.Active = true
	}

	rec := obj.Unique.TypeP.RecipieLookup[obj.Unique.SingleContent.TypeP.TypeI]
	if rec == nil {
		cwlog.DoLog(true, "Nil recipie")
		return
	}
	result := rec.ResultP[0]

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.TypeP.MachineSettings.KgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount--
	/* Subtract ore weight */
	obj.KGHeld -= obj.Unique.SingleContent.TypeP.KG

	/* Output result */
	obj.Outputs[0].Buf.Amount = result.KG

	/* Find and set result type, if needed */
	if obj.Outputs[0].Buf.TypeP != result {
		obj.Outputs[0].Buf.TypeP = result
	}
}

func slipRollerUpdate(obj *world.ObjData) {

	/* Get fuel */
	for _, fuel := range obj.FuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.TypeP.MachineSettings.MaxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
		continue
	}

	/* Check input */
	for _, input := range obj.Inputs {

		/* Input contains something */
		if input.Buf.Amount < 1 {
			continue
		}

		/* Contents is metal bar */
		if !input.Buf.TypeP.IsBar {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+(input.Buf.Amount) > obj.Unique.TypeP.MachineSettings.MaxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.TypeP != input.Buf.TypeP {
			obj.Unique.SingleContent.TypeP = input.Buf.TypeP
		}

		/* Add to weight */
		obj.KGHeld += (input.Buf.Amount * input.Buf.TypeP.KG)

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += (input.Buf.Amount * input.Buf.TypeP.KG)
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < 1 {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.TypeP.MachineSettings.KgFuelPerCycle {
		if obj.Active {
			obj.Active = false
		}
		return
	}

	if !obj.Active {
		obj.Active = true
	}

	rec := obj.Unique.TypeP.RecipieLookup[obj.Unique.SingleContent.TypeP.TypeI]
	if rec == nil {
		cwlog.DoLog(true, "Nil recipie")
		return
	}
	result := rec.ResultP[0]

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.TypeP.MachineSettings.KgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount--
	/* Subtract ore weight */
	obj.KGHeld -= obj.Unique.SingleContent.TypeP.KG

	/* Output result */
	obj.Outputs[0].Buf.Amount = result.KG

	/* Find and set result type, if needed */
	if obj.Outputs[0].Buf.TypeP != result {
		obj.Outputs[0].Buf.TypeP = result
	}
}
