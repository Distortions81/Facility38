package main

import (
	"GameTest/consts"
	"time"
)

func GLogic() {

	lastUpdate := time.Now()
	var ticks uint64 = 0

	for {
		if time.Since(lastUpdate) > consts.GameLogicRate {
			ticks++

			//fmt.Println("Tick:", ticks)
		}

		//Reduce busy waiting
		time.Sleep(consts.GameLogicSleep)
	}
}
