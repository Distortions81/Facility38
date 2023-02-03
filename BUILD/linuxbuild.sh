#!/bin/bash
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=50 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime" -o GameTest-$curTime-linux64
zip -9 GameTest-$curTime-linux64.zip GameTest-$curTime-linux64
rm GameTest-$curTime-linux64
