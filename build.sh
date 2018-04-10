#!/bin/bash
#########################################################
# Base Kubernetes Branch: kubernetes-1.9 65e0e7c723
# patch:
# change k8s.io/client-go to github.com/hyperhq/client-go
#########################################################

base_dir=$(cd $(dirname $0);pwd)
GOPATH=$(cd ../../../..;pwd)
export GOPATH=$base_dir/vendor:$GOPATH
go build -ldflags "-w" cmd/pi/pi.go

[ $? -eq 0 ] && echo "build pi done" || echo "build pi error"
