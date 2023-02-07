#!/bin/bash
rm html/main-bench.wasm.gz
rm html/main-bench.wasm

curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=50 GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -tags=ebitensinglethread -ldflags="-s -w -X main.buildTime=$curTime -X gv.WASMMode=true -X gv.UPSBench=true -X gv.LoadTest=true" -o html/main-bench.wasm
gzip -9 html/main-bench.wasm