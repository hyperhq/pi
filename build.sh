#!/bin/bash
#########################################################
# Base Kubernetes Branch: kubernetes-1.9 65e0e7c723
# patch:
# change k8s.io/client-go to github.com/hyperhq/client-go
#########################################################

export GOPATH=`pwd`/vendor:$GOPATH
go build -ldflags "-w" cmd/pi/pi.go

[ $? -eq 0 ] && echo "build pi done" || echo "build pi error"
