#!/bin/bash
path="BUILD/builds/"
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

# Check if an argument was passed in
if [ $# -eq 1 ]; then
  # Overwrite curTime with the argument
  curTime=$1
fi

go build -pgo=auto -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`

GOOS=linux GOARCH=amd64 go build -trimpath -gcflags=all="-B" -ldflags="-s -w -X main.buildTime=$curTime -X main.NoDebug=true" -o $path/Facility-38/Facility-38

cd $path
zip -9 Facility-38-$versionString-linux64.zip Facility-38/Facility-38
scp -P 5313 Facility-38-$versionString-linux64.zip dist@facility38.go-game.net:~/F38Auth/www/dl/
rm Facility-38/Facility-38