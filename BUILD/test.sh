#!/bin/bash

GOOS=linux GOARCH=amd64 go build -pgo=auto -trimpath -ldflags="-X main.buildTime=$curTime" -o Facility-38
./Facility38