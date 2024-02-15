#!/bin/bash

GOOS=linux GOMAXPROCS=1 GOARCH=amd64 go build -pgo=off -trimpath -gcflags=all="-B" -ldflags="-s -w" -tags=ebitensinglethread
./Facility38