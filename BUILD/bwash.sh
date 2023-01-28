#!/bin/bash
rm html/main.wasm.gz
rm html/main.wasm

curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=10 GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -tags=ebitensinglethread -ldflags="-s -w -X main.buildTime=$curTime" -o html/main.wasm
gzip -9 html/main.wasm