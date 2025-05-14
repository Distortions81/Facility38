#!/bin/bash

# exit when any command fails
set -e

#Make date string
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

#Build with time string
go build -ldflags="-X main.buildTime=$curTime"
#Get reply with full version string
versionString=`./Facility38 --version`

mv BUILD/builds/*.zip BUILD/builds/old-builds || true

#Push new version to auth server
../BuildUtilF38/BuildUtilF38 -build="$versionString" -path="../Facility38/data/txt/" -pass="invalid"

ssh -p 5313 dist@facility38.go-game.net "mv ~/F38Auth/www/dl/*.zip ~/F38Auth/www/dl/old-builds/" || true

#Build all with new key
bash BUILD/winbuild.sh $curTime
bash BUILD/linuxbuild.sh $curTime
bash BUILD/wasm.sh $curTime

#Make code backup for this version for debug ref
7z a -t7z BUILD/builds/code-ref/$versionString.7z *.go */*.go

#Let user know we are done.
echo "Build of '$versionString' complete, pushing build."

