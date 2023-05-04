#!/bin/bash
path="BUILD/builds"
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi

go build -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`

GOGC=100 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true" -o $path/Facility38-$versionString-linux64

cd $path
zip -9 Facility38-$versionString-linux64.zip Facility38-$versionString-linux64
scp -P 5313 Facility38-$versionString-linux64.zip dist@m45sci.xyz:~/F38Auth/www/dl/
rm Facility38-$versionString-linux64
