#!/usr/bin/env bash
## replace parameter in template with env variable

TMPL_DIR=template
OUT_DIR=.

for tmpl in `ls $TMPL_DIR`
do
	echo "convert yaml $TMPL_DIR/$tmpl"
	cat $TMPL_DIR/$tmpl |
	awk '$0 !~ /^\s*#.*$/' |
	sed 's/[ "]/\\&/g' |
	while read -r line;do
	    eval echo ${line}
	done > $OUT_DIR/$tmpl
done