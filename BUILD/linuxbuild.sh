#!/bin/bash
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi

GOGC=100 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true" -o BUILD/builds/Facility38-$curTime-linux64
zip -9 BUILD/builds/Facility38-$curTime-linux64.zip BUILD/builds/Facility38-$curTime-linux64
rm BUILD/builds/Facility38-$curTime-linux64
