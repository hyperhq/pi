#!/usr/bin/env bash

OS=$1
ARCH=$2
filename=""

base_dir=$(cd $(dirname $0);pwd)
cd $base_dir/..


if [ "$OS" == "" -o "$ARCH" == ""  ];then
	echo "can not get os info: OS($OS) ARCH($ARCH)"
	exit 1
fi

case $OS in
linux)
	filename=pi.$OS-$ARCH.tar.gz
	tar czvf pi.$OS-$ARCH.tar.gz pi
	;;
darwin)
	filename=pi.$OS-$ARCH.zip
	zip pi.$OS-$ARCH.zip pi
	;;
*)
	echo "os $OS not support"
	;;
esac

if [ $? -eq 0 ];then
	echo "$filename" packaged OK
else
	echo echo "$filename" packaged failed
	exit 1
fi
