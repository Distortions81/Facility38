#!/bin/bash
path="BUILD/builds/wasm"
echo "purging old builds..."
rm $path/main.wasm.gz
rm $path/main.wasm
rm $path/start.wasm

curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi

echo "compiling..."
go build -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`

GOOS=js GOMAXPROCS=1 GOARCH=wasm go build -tags=ebitensinglethread -pgo=auto -trimpath -gcflags=all="-B" -ldflags="-s -w -X main.buildTime=$curTime" -o $path/start.wasm

cd $path

echo "optimizing wasm..."
wasm-opt --enable-bulk-memory -O0 -o main.wasm start.wasm 
rm start.wasm
echo "compressing..."
gzip -9 main.wasm

echo "uploading build..."
scp -P 5313 main.wasm.gz dist@facility38.go-game.net:~/F38Auth/www/
mv main.wasm.gz ..