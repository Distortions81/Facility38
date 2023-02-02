package objects

import (
	"GameTest/consts"
	"GameTest/glob"
)

func minerUpdate(o *glob.ObjData) {
	if o.OutputBuffer.Amount == 0 {
		input := uint64((o.TypeP.MinerKGTock))

		o.OutputBuffer.Amount = input
		o.OutputBuffer.TypeP = *MatTypes[consts.MAT_COAL]
	}
}

func beltUpdate(o *glob.ObjData) {
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

func splitterUpdate(o *glob.ObjData) {
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

func boxUpdate(o *glob.ObjData) {

	for _, mat := range o.InputBuffer {
		if mat != nil && mat.Amount > 0 {
			if o.KGHeld+mat.Amount <= o.TypeP.CapacityKG {
				if o.Contents[mat.TypeP.TypeI] == nil {
					o.Contents[mat.TypeP.TypeI] = &glob.MatData{}
				}
				o.Contents[mat.TypeP.TypeI].Amount += mat.Amount
				o.KGHeld += mat.Amount
				o.Contents[mat.TypeP.TypeI].TypeP = mat.TypeP

				mat.Amount = 0
			}
		}
	}
}

func smelterUpdate(o *glob.ObjData) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func ironCasterUpdate(o *glob.ObjData) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func steamEngineUpdate(o *glob.ObjData) {
}
