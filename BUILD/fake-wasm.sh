#!/bin/bash
curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=10 GOOS=linux GOMAXPROCS=1 GOARCH=amd64 go build -tags=ebitensinglethread -ldflags="-s -w -X main.buildTime=$curTime -X gv.WASMMode=true" -o fakeWasm