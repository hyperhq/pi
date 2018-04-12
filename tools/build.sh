#!/bin/bash
#########################################################
# Base Kubernetes Branch: kubernetes-1.9 65e0e7c723
# patch:
# change k8s.io/client-go to github.com/hyperhq/client-go
#########################################################

base_dir=$(cd $(dirname $0);pwd)
cd $base_dir

GOPATH=$(cd ../../../../..;pwd)
export GOPATH=$base_dir/vendor:$GOPATH

LDFLAGS="$@"
LDFLAGS=${LDFLAGS:--w}

START=`date +"%s"`
cd ..
echo "LDFLAGS: $LDFLAGS"
go build -ldflags "$LDFLAGS"  cmd/pi/pi.go
END=`date +"%s"`

[ $? -eq 0 ] && echo "build pi done ($(($END - $START)) seconds)" || echo "build pi error"
