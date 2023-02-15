package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func toggleOverlay() {
	if world.ShowInfoLayer {
		world.ShowInfoLayer = false
		util.Chat("Info overlay is now off.")
	} else {
		world.ShowInfoLayer = true
		util.Chat("Info overlay is now on.")
	}
}

func initMiner(obj *world.ObjData) {
	if obj == nil {
		return
	}

	/* Init miner data if needed */
	if obj.MinerData == nil {
		obj.MinerData = &world.MinerDataType{}
	}

	foundRes := false
	/* Check for resources to mine */
	for p := 1; p < len(NoiseLayers); p++ {
		var h float32 = float32(math.Abs(float64(NoiseMap(float32(obj.Pos.X), float32(obj.Pos.Y), p))))

		/* We only mine solids */
		if !NoiseLayers[p].TypeP.IsSolid {
			continue
		}
		if h > 0.01 {
			obj.MinerData.Resources = append(obj.MinerData.Resources, h)
			obj.MinerData.ResourcesType = append(obj.MinerData.ResourcesType, NoiseLayers[p].TypeI)
			obj.MinerData.ResourcesCount++
			foundRes = true
		}
	}

	/* Nothing to mine here, kill all the events for this miner */
	if !foundRes {

		/* Let user know of this */
		oPos := util.CenterXY(obj.Pos)
		util.ChatDetailed(fmt.Sprintf("%v at (%v,%v): No resources to mine here!", obj.TypeP.Name, oPos.X, oPos.Y),
			world.ColorRed, time.Minute)

		obj.Blocked = true
		obj.Active = false
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)

		return /* Stop here */
	}

	/* Init ResourcesMined if needed */
	if obj.Parent.ResourcesMined == nil {
		obj.Parent.ResourcesMined = []*world.ResUsedData{}
	}

	/*
	 * Lets check if this chunk and position has ResourcesMined data.
	 * If it does not:
	 *
	 * Init our position in the ResourcesMined list now
	 * so that concurrent access is possible later.
	 * Also link to the data in obj's MinerData, to avoid lookups
	 */

	found := false
	for i, item := range obj.Parent.ResourcesMined {
		if item.Pos == obj.Pos {
			//Add a link to our position for access later
			obj.MinerData.TotalMined = obj.Parent.ResourcesMined[i]
			found = true
			break
		}
	}

	/* Not found, add */
	if !found {
		insert := &world.ResUsedData{Pos: obj.Pos}
		obj.Parent.ResourcesMined = append(obj.Parent.ResourcesMined, insert)

		//Add a link to our position for access later
		obj.MinerData.TotalMined = insert
	}

}

func minerUpdate(obj *world.ObjData) {

	/* Cycle all ports */
	for p, port := range obj.Ports {
		/* Fuel input */
		if port.PortDir == gv.PORT_INPUT {

			/* Valid? */
			if port.Buf.TypeP == nil {
				continue
			}

			/* Is it fuel? */
			if port.Buf.TypeP.TypeI != gv.MAT_COAL {

				/* Will it over fill us? */
				if obj.KGFuel+port.Buf.Amount <= obj.TypeP.MaxFuelKG {

					/* Eat the fuel and increase fuel kg */
					obj.KGFuel += port.Buf.Amount
					obj.Ports[p].Buf.Amount = 0
				}
			}
		}

		/* Output full? */
		if port.Buf.Amount != 0 {
			obj.Blocked = true
			obj.Active = false
			continue
		}

		/* Then we are not blocked */
		obj.Blocked = false

		/* Turn on active status */
		obj.Active = true

		/* Are we ready to output yet? */
		if obj.TickCount < obj.TypeP.Interval {
			/* Increment timer */
			obj.TickCount++
			continue
		}

		/* Randomly pick a material from the list */
		pick := rand.Intn(int(obj.MinerData.ResourcesCount))

		/* Calculate how much material */
		amount := obj.TypeP.KgMineEach * float32(obj.MinerData.Resources[pick])
		kind := MatTypes[obj.MinerData.ResourcesType[pick]]

		/* Are we are mining coal? */
		if obj.MinerData.ResourcesType[pick] == gv.MAT_COAL &&
			obj.KGFuel+amount <= obj.TypeP.MaxFuelKG {

			/* If we need fuel, fuel ourselves */
			obj.KGFuel += amount
			continue
		}
		if obj.KGFuel < obj.TypeP.KgFuelEach {
			/* Not enough fuel */
			continue
		}

		/* Otherwise output the material */
		obj.Ports[obj.Dir].Buf.Amount = amount
		obj.Ports[obj.Dir].Buf.TypeP = kind
		obj.Ports[obj.Dir].Buf.Rot = uint8(rand.Intn(3))

		/* Burn fuel */
		obj.KGFuel -= obj.TypeP.KgFuelEach

		obj.TickCount = 0

		//We should remove ourselves here if we run out of ore
	}
}

func beltUpdateInter(obj *world.ObjData) {

	for p, port := range obj.Ports {
		/* Valid? */
		if port == nil {
			continue
		}

		/* Input? */
		if port.PortDir != gv.PORT_INPUT {
			continue
		}

		/* Valid output on other side? */
		odir := util.ReverseDirection(uint8(p))
		if obj.Ports[odir] == nil {
			continue
		}

		/* Do we have input and is output is empty */
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
	worked := false
	for x := 0; x < 4; x++ {
		dir = util.RotCW(dir)

		/* Is this an input? */
		if obj.Ports[dir].PortDir != gv.PORT_INPUT {
			continue
		}

		/* Does the input contain anything? */
		if obj.Ports[dir].Buf.Amount == 0 {
			continue
		} else {
			worked = true
			obj.Active = true
			obj.Ports[obj.Dir].Buf.Amount = obj.Ports[dir].Buf.Amount
			obj.Ports[obj.Dir].Buf.TypeP = obj.Ports[dir].Buf.TypeP
			obj.Ports[obj.Dir].Buf.Rot = obj.Ports[dir].Buf.Rot
			obj.Ports[dir].Buf.Amount = 0
			obj.LastUsedInput = dir
			break /* Stop */
		}
	}
	if !worked {
		obj.Active = false
	}
}

func fuelHopperUpdate(obj *world.ObjData) {

	/* Valid port? */
	if obj.Ports[obj.Dir] == nil {
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

		/* Is this the output? */
		if obj.Ports[dir].PortDir != gv.PORT_OUTPUT {
			continue
		}

		/* Is the port empty? */
		if obj.Ports[dir].Buf.Amount != 0 {
			obj.Active = false
			continue
		} else {
			/* Otherwise output */
			obj.Ports[dir].Buf.Amount = obj.Ports[input].Buf.Amount
			obj.Ports[dir].Buf.TypeP = obj.Ports[input].Buf.TypeP
			obj.Ports[dir].Buf.Rot = obj.Ports[input].Buf.Rot
			obj.Ports[input].Buf.Amount = 0
			obj.LastUsedOutput = dir
			obj.Active = true
			break
			/* End */
		}

	}

}

func boxUpdate(obj *world.ObjData) {
	for p, port := range obj.Ports {

		if port.PortDir == gv.PORT_INPUT {

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
			obj.Contents[port.Buf.TypeP.TypeI].Amount += obj.Ports[p].Buf.Amount
			obj.Contents[port.Buf.TypeP.TypeI].TypeP = obj.Ports[p].Buf.TypeP
			obj.KGHeld += port.Buf.Amount
			obj.Ports[p].Buf.Amount = 0
			continue
		}

		//Unloader goes here
	}
}

func smelterUpdate(obj *world.ObjData) {

	for p, port := range obj.Ports {

		/* Valid? */
		if port == nil {
			continue
		}

		/* Input? */
		if port.PortDir == gv.PORT_INPUT {

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
