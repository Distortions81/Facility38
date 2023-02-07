#!/bin/bash
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=100 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime -X gv.UPSBench=true -X gv.LoadTest=true" -o GameTest-$curTime-linux64-bench
zip -9 GameTest-$curTime-linux64-bench.zip GameTest-$curTime-linux64-bench
rm GameTest-$curTime-linux64-bench
