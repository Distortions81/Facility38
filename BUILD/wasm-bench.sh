#!/bin/bash
rm BUILD/builds/wasm/main-bench.wasm.gz
rm BUILD/builds/wasm/main-bench.wasm

curTime=`date -u '+%Y%m%d%H%M%S'`

GOGC=50 GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -tags=ebitensinglethread -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug -X main.UPSBench=true -X main.LoadTest=true" -o BUILD/builds/wasm/main-bench.wasm
gzip -9 BUILD/builds/wasm/main-bench.wasm
rm BUILD/builds/wasm/main-bench.wasm