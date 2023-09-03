#!/bin/bash

GOOS=linux GOARCH=amd64 go build -pgo=auto
./Facility38