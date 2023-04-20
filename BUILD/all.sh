#!/bin/bash

curTime=`date -u '+%Y%m%d%H%M%S'`

bash BUILD/winbuild.sh $curTime
bash BUILD/linuxbuild.sh $curTime
bash BUILD/wasm.sh $curTime