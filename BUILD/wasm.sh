#!/bin/bash
path="BUILD/builds/wasm"
rm BUILD/main.wasm.gz
rm $path/main.wasm
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi

go build -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`



GOGC=256 GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -trimpath -tags=ebitensinglethread -gcflags=all="-l -B" -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true -X main.WASMMode=true" -o $path/main.wasm

cd $path
gzip -9 main.wasm
scp -P 5313 main.wasm.gz dist@facility38.go-game.net:~/F38Auth/www/
mv main.wasm.gz ..
