#!/bin/bash

GOOS=linux GOARCH=amd64 go build -pgo=auto -trimpath -gcflags=all="-B" -ldflags="-s -w"
./Facility38