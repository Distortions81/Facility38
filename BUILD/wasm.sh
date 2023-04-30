#!/bin/bash
rm BUILD/builds/wasm/main.wasm.gz
rm BUILD/builds/wasm/main.wasm

curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`
go build -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  versionString=$1
fi

GOGC=100 GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -trimpath -tags=ebitensinglethread -ldflags="-s -w -X main.buildTime=$versionString -X main.NoDebug=true -X main.WASMMode=true" -o BUILD/builds/wasm/main.wasm
gzip -9 BUILD/builds/wasm/main.wasm
rm BUILD/builds/wasm/main.wasm