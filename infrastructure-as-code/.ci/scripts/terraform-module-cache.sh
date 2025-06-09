#! /usr/bin/env bash

dir=$1

for dir in $(ls -d $dir/*); do
	pushd $dir
	terraform init -backend=false
	popd
done
tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
zip terraform.zip -r "$tdir" && rm -rf "$tdir"
