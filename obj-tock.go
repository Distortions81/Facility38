package main

import (
	"math/rand"
)

func minerUpdate(obj *ObjData) {
	defer reportPanic("minerUpdate")

	/* Get fuel */
	for p, port := range obj.fuelIn {
		/* Will it over fill us? */
		if port.Buf.Amount > 0 &&
			obj.Unique.KGFuel+port.Buf.Amount <= obj.Unique.typeP.machineSettings.maxFuelKG {

			/* Eat the fuel */
			obj.Unique.KGFuel += port.Buf.Amount
			obj.fuelIn[p].Buf.Amount = 0
		}
	}

	if obj.Unique.KGFuel < obj.Unique.typeP.machineSettings.kgFuelPerCycle {
		/* Not enough fuel, exit */
		if obj.active {
			obj.active = false
		}
		return
	}

	if obj.numOut == 0 {
		return
	}

	/* Cycle through available materials */
	var pick uint8 = 0

	if obj.MinerData.resourcesCount > 1 {
		if obj.MinerData.lastUsed < (obj.MinerData.resourcesCount - 1) {
			obj.MinerData.lastUsed++
		} else {
			obj.MinerData.lastUsed = 0
		}
		pick = obj.MinerData.lastUsed
	}

	/* Calculate how much material */
	if obj.MinerData.resourcesCount == 0 {
		return
	}
	amount := obj.Unique.typeP.machineSettings.kgPerCycle * float32(obj.MinerData.resources[pick])
	kind := matTypes[obj.MinerData.resourcesType[pick]]

	/* Stop if the amount is extremely small, zero or negative */
	if amount < 0.001 {
		return
	}

	/* Set as actively working */
	if !obj.active {
		obj.active = true
	}

	/* Tally the amount taken as well as the type */
	obj.Tile.minerData.mined[pick] += amount

	/* Output the material */
	obj.outputs[0].Buf.Amount = amount
	if obj.outputs[0].Buf.typeP != kind {
		obj.outputs[0].Buf.typeP = kind
	}
	obj.outputs[0].Buf.Rot = uint8(rand.Intn(3))

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.typeP.machineSettings.kgFuelPerCycle

}

func beltUpdateOver(obj *ObjData) {
	defer reportPanic("beltUpdateOver")

	/* Underpass */
	if obj.beltOver.underIn != nil && obj.beltOver.underOut != nil {
		if obj.beltOver.underOut.obj != nil && !obj.beltOver.underOut.obj.blocked {
			if obj.beltOver.underIn.Buf.Amount != 0 && obj.beltOver.underOut.Buf.Amount == 0 {
				*obj.beltOver.underOut.Buf, *obj.beltOver.underIn.Buf = *obj.beltOver.underIn.Buf, *obj.beltOver.underOut.Buf
			}
		}
	}

	/* Overpass to OverOut */
	if obj.beltOver.overOut != nil && obj.beltOver.middle != nil {
		if obj.beltOver.overOut.obj != nil {
			if obj.beltOver.middle.Amount != 0 && obj.beltOver.overOut.Buf.Amount == 0 {
				*obj.beltOver.overOut.Buf, *obj.beltOver.middle = *obj.beltOver.middle, *obj.beltOver.overOut.Buf
			}
		}
	}

	/* OverIn to Overpass */
	if obj.beltOver.overIn != nil && obj.beltOver.middle != nil {
		if obj.beltOver.overIn.Buf.Amount != 0 && obj.beltOver.middle.Amount == 0 {
			*obj.beltOver.middle, *obj.beltOver.overIn.Buf = *obj.beltOver.overIn.Buf, *obj.beltOver.middle
		}
	}

}

func beltUpdate(obj *ObjData) {
	defer reportPanic("beltUpdate")

	if obj.numIn > 1 {
		if obj.LastInput == (obj.numIn - 1) {
			obj.LastInput = 0
		} else {
			obj.LastInput++
		}
	}

	/* Does the input contain anything? */
	if obj.numOut > 0 && obj.numIn > 0 {
		if obj.inputs[obj.LastInput].Buf.Amount > 0 &&
			obj.outputs[0].Buf.Amount == 0 &&
			obj.outputs[0].obj != nil &&
			!obj.outputs[0].obj.blocked {
			/* Good to go, swap pointers */
			*obj.outputs[0].Buf, *obj.inputs[obj.LastInput].Buf = *obj.inputs[obj.LastInput].Buf, *obj.outputs[0].Buf
		}
	}
}

func fuelHopperUpdate(obj *ObjData) {
	defer reportPanic("fuelHopperUpdate")

	for i, input := range obj.inputs {

		/* Does input contain anything? */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Is input solid? */
		if !input.Buf.typeP.isSolid {
			continue
		}

		/* Is input fuel? */
		if !input.Buf.typeP.isFuel {
			continue
		}

		/* Do we have room for it? */
		if (obj.Unique.KGFuel + input.Buf.Amount) < obj.Unique.typeP.machineSettings.maxFuelKG {
			obj.Unique.KGFuel += input.Buf.Amount
			obj.inputs[i].Buf.Amount = 0
			break
		}
	}

	if obj.Unique.KGFuel > 0 {
		if !obj.active {
			obj.active = true
		}
	} else {
		if obj.active {
			obj.active = false
		}
	}

	/* Grab destination object */
	if obj.Unique.KGFuel > (obj.Unique.typeP.machineSettings.kgHopperMove + obj.Unique.typeP.machineSettings.kgFuelPerCycle) {
		for _, output := range obj.fuelOut {
			output.Buf.Amount = obj.Unique.typeP.machineSettings.kgHopperMove
			obj.Unique.KGFuel -= (obj.Unique.typeP.machineSettings.kgHopperMove + obj.Unique.typeP.machineSettings.kgFuelPerCycle)
			break
		}
	}
}

func loaderUpdate(obj *ObjData) {
	defer reportPanic("loaderUpdate")

	for i, input := range obj.inputs {
		if input.Buf.Amount == 0 {
			continue
		}
		if obj.numOut == 0 || obj.outputs[0].Buf.Amount != 0 {
			continue
		}
		*obj.outputs[0].Buf, *obj.inputs[i].Buf = *obj.inputs[i].Buf, *obj.outputs[0].Buf
		break
	}
}

func splitterUpdate(obj *ObjData) {
	defer reportPanic("splitterUpdate")

	if obj.numIn > 0 && obj.inputs[0].Buf.Amount > 0 {
		if obj.numOut > 0 {
			if obj.LastOutput >= (obj.numOut - 1) {
				obj.LastOutput = 0
			} else {
				obj.LastOutput++
			}
		} else {
			return
		}

		if obj.outputs[obj.LastOutput].Buf.Amount == 0 {
			/* Good to go, swap pointers */
			*obj.inputs[0].Buf, *obj.outputs[obj.LastOutput].Buf = *obj.outputs[obj.LastOutput].Buf, *obj.inputs[0].Buf
			return
		}
	}
}

func boxUpdate(obj *ObjData) {
	defer reportPanic("boxUpdate")

	for p, port := range obj.inputs {
		if port.Buf.typeP == nil {
			continue
		}

		/* Will the input fit? */
		if obj.KGHeld+port.Buf.Amount > obj.Unique.typeP.machineSettings.maxContainKG {
			continue
		}

		/* Init content type if needed */
		if obj.Unique.Contents.mats[port.Buf.typeP.typeI] == nil {
			obj.Unique.Contents.mats[port.Buf.typeP.typeI] = &MatData{}
		}

		/* Add to contents */
		obj.Unique.Contents.mats[port.Buf.typeP.typeI].Amount += obj.inputs[p].Buf.Amount
		obj.Unique.Contents.mats[port.Buf.typeP.typeI].typeP = matTypes[port.Buf.typeP.typeI]
		obj.KGHeld += port.Buf.Amount
		obj.inputs[p].Buf.Amount = 0
		continue

	}
}

func smelterUpdate(obj *ObjData) {
	defer reportPanic("smelterUpdate")

	/* Get fuel */
	for _, fuel := range obj.fuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.typeP.machineSettings.maxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
	}

	/* Check input */
	for _, input := range obj.inputs {

		/* Input contains something */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Input is ore */
		if !input.Buf.typeP.isOre {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+input.Buf.Amount > obj.Unique.typeP.machineSettings.maxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.typeP != input.Buf.typeP {
			if obj.Unique.SingleContent.Amount > 0 {
				obj.Unique.SingleContent.typeP = matTypes[MAT_MIX_ORE]
			} else {
				obj.Unique.SingleContent.typeP = input.Buf.typeP
			}
		}

		/* Add to weight */
		obj.KGHeld += input.Buf.Amount

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += input.Buf.Amount
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < obj.Unique.typeP.machineSettings.kgPerCycle {
		if obj.active {
			obj.active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.typeP.machineSettings.kgFuelPerCycle {
		if obj.active {
			obj.active = false
		}
		return
	}

	if !obj.active {
		obj.active = true
	}

	/* Look up material */
	rec := obj.Unique.typeP.recipeLookup[obj.Unique.SingleContent.typeP.typeI]
	if rec == nil {
		doLog(true, "Nil recipe")
		return
	}
	result := rec.resultP[0]

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.typeP.machineSettings.kgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount -= obj.Unique.typeP.machineSettings.kgPerCycle
	/* Subtract ore weight */
	obj.KGHeld -= obj.Unique.typeP.machineSettings.kgPerCycle

	/* Output result */
	if obj.numOut > 0 {
		obj.outputs[0].Buf.Amount = obj.Unique.typeP.machineSettings.kgPerCycle

		/* Find and set result type, if needed */
		if obj.outputs[0].Buf.typeP != result {
			obj.outputs[0].Buf.typeP = result
		}
	}
}

func casterUpdate(obj *ObjData) {
	defer reportPanic("casterUpdate")

	/* Get fuel */
	for _, fuel := range obj.fuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.typeP.machineSettings.maxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
		continue
	}

	/* Check input */
	for _, input := range obj.inputs {

		/* Input contains something */
		if input.Buf.Amount == 0 {
			continue
		}

		/* Contents are shot */
		if !input.Buf.typeP.isShot {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+input.Buf.Amount > obj.Unique.typeP.machineSettings.maxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.typeP != input.Buf.typeP {
			obj.Unique.SingleContent.typeP = input.Buf.typeP
		}

		/* Add to weight */
		obj.KGHeld += input.Buf.Amount

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += input.Buf.Amount
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < 1 {
		if obj.active {
			obj.active = false
		}
		return
	}

	/* Process ores */
	/* Is there enough ore to process? */
	rec := obj.Unique.typeP.recipeLookup[obj.Unique.SingleContent.typeP.typeI]
	if rec == nil {
		doLog(true, "Nil recipe")
		return
	}
	result := rec.resultP[0]

	if obj.Unique.SingleContent.Amount < result.kg {
		if obj.active {
			obj.active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.typeP.machineSettings.kgFuelPerCycle {
		if obj.active {
			obj.active = false
		}
		return
	}

	if !obj.active {
		obj.active = true
	}

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.typeP.machineSettings.kgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount -= result.kg
	/* Subtract ore weight */
	obj.KGHeld -= result.kg

	/* Output result */
	obj.outputs[0].Buf.Amount = result.kg

	/* Find and set result type, if needed */
	if obj.outputs[0].Buf.typeP != result {
		obj.outputs[0].Buf.typeP = result
	}
}

func rodCasterUpdate(obj *ObjData) {
	defer reportPanic("rodCasterUpdate")

	/* Get fuel */
	for _, fuel := range obj.fuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.typeP.machineSettings.maxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
		continue
	}

	/* Check input */
	for _, input := range obj.inputs {

		/* Input contains something */
		if input.Buf.Amount < 1 {
			continue
		}

		/* Contents is metal bar */
		if !input.Buf.typeP.isBar {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+(input.Buf.Amount) > obj.Unique.typeP.machineSettings.maxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.typeP != input.Buf.typeP {
			obj.Unique.SingleContent.typeP = input.Buf.typeP
		}

		/* Add to weight */
		obj.KGHeld += (input.Buf.Amount * input.Buf.typeP.kg)

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += (input.Buf.Amount * input.Buf.typeP.kg)
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < 1 {
		if obj.active {
			obj.active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.typeP.machineSettings.kgFuelPerCycle {
		if obj.active {
			obj.active = false
		}
		return
	}

	if !obj.active {
		obj.active = true
	}

	rec := obj.Unique.typeP.recipeLookup[obj.Unique.SingleContent.typeP.typeI]
	if rec == nil {
		doLog(true, "Nil recipe")
		return
	}
	result := rec.resultP[0]

	/* Burn fuel */
	obj.Unique.KGFuel -= obj.Unique.typeP.machineSettings.kgFuelPerCycle

	/* Subtract ore */
	obj.Unique.SingleContent.Amount--
	/* Subtract ore weight */
	obj.KGHeld -= obj.Unique.SingleContent.typeP.kg

	/* Output result */
	obj.outputs[0].Buf.Amount = result.kg

	/* Find and set result type, if needed */
	if obj.outputs[0].Buf.typeP != result {
		obj.outputs[0].Buf.typeP = result
	}
}

func slipRollerUpdate(obj *ObjData) {
	defer reportPanic("slipRollerUpdate")

	/* Get fuel */
	for _, fuel := range obj.fuelIn {

		/* Will the fuel fit? */
		if obj.Unique.KGFuel+fuel.Buf.Amount > obj.Unique.typeP.machineSettings.maxFuelKG {
			continue
		}

		obj.Unique.KGFuel += fuel.Buf.Amount
		fuel.Buf.Amount = 0
		continue
	}

	/* Check input */
	for _, input := range obj.inputs {

		/* Input contains something */
		if input.Buf.Amount < 1 {
			continue
		}

		/* Contents is metal bar */
		if !input.Buf.typeP.isBar {
			continue
		}

		/* Contents will fit */
		if obj.KGHeld+(input.Buf.Amount) > obj.Unique.typeP.machineSettings.maxContainKG {
			continue
		}

		/* Set type if needed */
		if obj.Unique.SingleContent.typeP != input.Buf.typeP {
			obj.Unique.SingleContent.typeP = input.Buf.typeP
		}

		/* Add to weight */
		obj.KGHeld += (input.Buf.Amount * input.Buf.typeP.kg)

		/* Add input to contents */
		obj.Unique.SingleContent.Amount += (input.Buf.Amount * input.Buf.typeP.kg)
		input.Buf.Amount = 0
	}

	/* Is there enough ore to process? */
	if obj.Unique.SingleContent.Amount < 1 {
		if obj.active {
			obj.active = false
		}
		return
	}

	/* Do we have enough fuel? */
	if obj.Unique.KGFuel < obj.Unique.typeP.machineSettings.kgFuelPerCycle {
		if obj.active {
			obj.active = false
		}
		return
	}

	if !obj.active {
		obj.active = true
	}

	rec := obj.Unique.typeP.recipeLookup[obj.Unique.SingleContent.typeP.typeI]
	if rec == nil {
		doLog(true, "Nil recipe")
		return
	}

	if obj.numOut > 0 {
		result := rec.resultP[0]

		/* Burn fuel */
		obj.Unique.KGFuel -= obj.Unique.typeP.machineSettings.kgFuelPerCycle

		/* Subtract ore */
		obj.Unique.SingleContent.Amount--
		/* Subtract ore weight */
		obj.KGHeld -= obj.Unique.SingleContent.typeP.kg

		/* Output result */
		obj.outputs[0].Buf.Amount = result.kg

		/* Find and set result type, if needed */
		if obj.outputs[0].Buf.typeP != result {
			obj.outputs[0].Buf.typeP = result
		}
	}
}
