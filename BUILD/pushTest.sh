#!/bin/bash

# exit when any command fails
set -e

#Make date string
curTime=`date -u '+%Y-%m-%d-%H-%M-%S'`

#Build with time string
go build -ldflags="-X main.buildTime=$curTime"
#Get reply with full version string
versionString=`./Facility38 --version`

#Push new version to auth server
../BuildUtilF38/BuildUtilF38 -build="$versionString" -path="../Facility38/data/txt/" -pass="invalid"
