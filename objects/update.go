package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"time"
)

func MinerUpdate(o *glob.WObject, tickNow time.Time) {

	if o.OutputBuffer.Amount == 0 {
		input := uint64((o.TypeP.MinerKGTock))

		o.OutputBuffer.Amount = input
		o.OutputBuffer.TypeI = consts.MAT_COAL
		o.OutputBuffer.TypeP = *MatTypes[consts.MAT_COAL]
		o.OutputBuffer.TweenStamp = tickNow

		//fmt.Println("Miner: ", o.TypeP.Name, " output: ", input)
	}
}

func SmelterUpdate(obj *glob.WObject, tickNow time.Time) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(obj *glob.WObject, tickNow time.Time) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func BeltUpdate(obj *glob.WObject, tickNow time.Time) {
	if obj.OutputBuffer.Amount == 0 {
		for src, mat := range obj.InputBuffer {
			if mat.Amount > 0 {
				obj.OutputBuffer.TweenStamp = tickNow
				obj.OutputBuffer.Amount = mat.Amount
				obj.OutputBuffer.TypeI = mat.TypeI
				obj.OutputBuffer.TypeP = mat.TypeP
				obj.InputBuffer[src].Amount = 0
				//fmt.Println(obj.TypeP.Name, " moved: ", mat.Amount)
			}
		}
	}

}

func SteamEngineUpdate(obj *glob.WObject, tickNow time.Time) {
}

func BoxUpdate(obj *glob.WObject, tickNow time.Time) {

	for src, mat := range obj.InputBuffer {
		if mat.Amount > 0 {
			if obj.KGHeld+mat.Amount <= obj.TypeP.CapacityKG {
				if obj.Contents[mat.TypeI] == nil {
					obj.Contents[mat.TypeI] = &glob.MatData{}
				}
				obj.Contents[mat.TypeI].Amount += mat.Amount
				obj.KGHeld += mat.Amount
				obj.Contents[mat.TypeI].TypeI = mat.TypeI
				obj.Contents[mat.TypeI].TypeP = mat.TypeP

				obj.InputBuffer[src].Amount = 0
				//fmt.Println(MatTypes[mat.TypeI].Name, " input: ", mat.Amount)
			}
		}
	}
}
