package objects

import (
	"GameTest/consts"
	"GameTest/glob"
)

func MinerUpdate(o *glob.WObject) {

	if o.OutputBuffer.Amount == 0 {
		input := uint64((o.TypeP.MinerKGTock))

		o.OutputBuffer.Amount = input
		o.OutputBuffer.TypeP = *MatTypes[consts.MAT_COAL]

		//fmt.Println("Miner: ", o.TypeP.Name, " output: ", input)
	}
}

func BeltUpdate(o *glob.WObject) {
	if o.OutputBuffer.Amount == 0 {
		for src, mat := range o.InputBuffer {
			if mat != nil && mat.Amount > 0 {
				o.OutputBuffer.Amount = mat.Amount
				o.OutputBuffer.TypeP = mat.TypeP
				o.InputBuffer[src].Amount = 0
				//fmt.Println(obj.TypeP.Name, " moved: ", mat.Amount)
			}
		}
	}

}

func SplitterUpdate(o *glob.WObject) {
	if o.OutputBuffer.Amount == 0 {
		for src, mat := range o.InputBuffer {
			if mat != nil && mat.Amount > 0 {
				o.OutputBuffer.Amount = mat.Amount
				o.OutputBuffer.TypeP = mat.TypeP
				o.InputBuffer[src].Amount = 0
			}
		}
	}
}

func BoxUpdate(o *glob.WObject) {

	for src, mat := range o.InputBuffer {
		if mat != nil && mat.Amount > 0 {
			if o.KGHeld+mat.Amount <= o.TypeP.CapacityKG {
				if o.Contents[mat.TypeP.TypeI] == nil {
					o.Contents[mat.TypeP.TypeI] = &glob.MatData{}
				}
				o.Contents[mat.TypeP.TypeI].Amount += mat.Amount
				o.KGHeld += mat.Amount
				o.Contents[mat.TypeP.TypeI].TypeP = mat.TypeP

				o.InputBuffer[src].Amount = 0
			}
		}
	}
}

func SmelterUpdate(o *glob.WObject) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(o *glob.WObject) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func SteamEngineUpdate(o *glob.WObject) {
}
