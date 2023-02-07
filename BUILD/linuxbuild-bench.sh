#!/bin/bash
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=100 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug -X main.UPSBench=true -X main.LoadTest=true" -o BUILD/builds/GameTest-$curTime-linux64-bench
zip -9 BUILD/builds/GameTest-$curTime-linux64-bench.zip BUILD/builds/GameTest-$curTime-linux64-bench
rm BUILD/builds/GameTest-$curTime-linux64-bench
