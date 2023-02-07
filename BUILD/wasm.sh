#!/bin/bash
rm html/main.wasm.gz
rm html/main.wasm

curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=50 GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -tags=ebitensinglethread -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug -X gv.WASMMode=true" -o BUILD/builds/wasm/main.wasm
gzip -9 BUILD/builds/wasm/main.wasm
rm BUILD/builds/wasm/main.wasm