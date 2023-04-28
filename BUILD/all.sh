#!/bin/bash

curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

bash BUILD/winbuild.sh $curTime
bash BUILD/linuxbuild.sh $curTime
bash BUILD/wasm.sh $curTime

7z a -t7z BUILD/builds/code-ref/$curTime.7z *.go */*.go