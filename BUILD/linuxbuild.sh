#!/bin/bash
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=100 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug" -o BUILD/builds/GameTest-$curTime-linux64
zip -9 BUILD/builds/GameTest-$curTime-linux64.zip BUILD/builds/GameTest-$curTime-linux64
rm BUILD/builds/GameTest-$curTime-linux64
