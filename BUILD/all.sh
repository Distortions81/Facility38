#!/bin/bash

# exit when any command fails
set -e

curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`
go build -ldflags="-X main.buildTime=$curTime"
versionString=`./Facility38 --version`

mv BUILD/builds/*.zip BUILD/builds/old-builds || true

bash BUILD/winbuild.sh $curTime
bash BUILD/linuxbuild.sh $curTime
bash BUILD/wasm.sh $curTime

7z a -t7z BUILD/builds/code-ref/$versionString.7z *.go */*.go
echo "Build of '$versionString' complete."
