#!/bin/bash
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=50 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true -X main.WASMMode=true" -o BUILD/builds/fakeWasm.exe