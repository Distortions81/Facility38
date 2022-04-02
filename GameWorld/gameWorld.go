package GameWorld

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
	"time"
)

var LastWorldUpdate time.Time
var WorldTicks int

func Update() {

	if time.Since(LastWorldUpdate) > (consts.WorldUpdateMS * time.Millisecond) {

		LastWorldUpdate = time.Now()
		WorldTicks++

		//Main world update
		//fmt.Println("World Update:", WorldTicks)

		for _, chunk := range glob.WorldMap {
			for okey, obj := range chunk.MObj {
				//Ignore empty objects
				if obj.Type != glob.ObjTypeNone {
					//Async object login
					if glob.ObjTypes[obj.Type].HasAsync {
						go UpdateObjectAsync(okey, &obj)
					}

					//In-Order object logics
					fmt.Println("Would update object:", okey)

				} else {
					fmt.Println("Empty object encountered.")
				}
			}
		}
	}
}

func UpdateObjectAsync(pos glob.Position, obj *glob.MObj) {

}
