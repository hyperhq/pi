#!/usr/bin/env bash
# Author: Stefan Buck
# License: MIT
# https://gist.github.com/stefanbuck/ce788fee19ab6eb0b4447a85fc99f447
#
#
# This script accepts the following parameters:
#
# * owner
# * repo
# * tag
# * filename
# * github_api_token
#
# Script to upload a release asset using the GitHub API v3.
#
# Example:
#
# upload-github-release-asset.sh github_api_token=TOKEN owner=stefanbuck repo=playground tag=v0.1.0 filename=./build.zip
#

# Check dependencies.
set -e

base_dir=$(cd $(dirname $0);pwd)
cd $base_dir/..

xargs=$(which gxargs || which xargs)

# Validate settings.
[ "$TRACE" ] && set -x

CONFIG=$@

for line in $CONFIG; do
  eval "$line"
done

if [ "$github_api_token" == "" ];then
	github_api_token=$GITHUB_API_TOKEN
fi

# Define variables.
GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$owner/$repo"
GH_TAGS="$GH_REPO/releases/tags/$tag"
AUTH="Authorization: token $github_api_token"
WGET_ARGS="--content-disposition --auth-no-challenge --no-cookie"
CURL_ARGS="-LJO#"

if [[ "$tag" == "latest" ]]; then
	echo "can not use latest as tag"
	exit 1
fi


if [ "$os" == "linux" ];then
	FILENAME=pi.$os-$arch.tar.gz
else
	FILENAME=pi.$os-$arch.zip
fi


# Validate token.
curl -o /dev/null -sH "$AUTH" $GH_REPO || { echo "Error: Invalid repo, token or network issue!";  exit 1; }

# Read asset tags.
response=$(curl -sH "$AUTH" $GH_TAGS)

# Get ID of the asset based on given filename.
eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
[ "$id" ] || { echo "Error: Failed to get release id for tag: $tag"; echo "$response" | awk 'length($0)<100' >&2; exit 1; }


echo "Check exit asset... "

# List assets
GH_ASSET="https://api.github.com/repos/$owner/$repo/releases/$id/assets"
assets=$(curl -s "$GITHUB_OAUTH_BASIC" -H "Authorization: token $github_api_token" $GH_ASSET)
asset_name=$(echo $assets | jq -r '. | map(select(.name=="'$FILENAME'")) | .[0].name')

if [ "${asset_name}" != "null" ];then
	asset_id=$(echo $assets | jq -r '. | map(select(.name=="'$FILENAME'")) | .[0].id')
	GH_ASSET="https://api.github.com/repos/$owner/$repo/releases/assets/$asset_id"
	echo "> file $FILENAME already exists, delete the old asset $asset_name(id:$asset_id) first"
	curl -s "$GITHUB_OAUTH_BASIC" -X DELETE -H "Authorization: token $github_api_token" $GH_ASSET
	if [ $? -eq 0 ];then
		echo "> old asset $asset_name(id:$asset_id) deleted"
	fi
else
	echo "> there is no exist $FILENAME"
fi



# Upload asset
echo "Uploading new asset... (tag:$tag, filename:$FILENAME) "

# Construct url
GH_ASSET="https://uploads.github.com/repos/$owner/$repo/releases/$id/assets?name=$(basename $FILENAME)"

START=`date +"%s"`
set +e
curl -# -L -o "$GITHUB_OAUTH_BASIC" -X POST --data-binary @"$FILENAME" -H "Authorization: token $github_api_token" -H "Content-Type: application/octet-stream" $GH_ASSET
if [ $? -eq 0 -o $? -eq 23 ];then
	END_UPLOAD=`date +"%s"`
	if [ "$NEED_TEST_DOWNLOAD" == "true" ];then
		echo "start test download"
		DOWNLOAD_URL="https://github.com/hyperhq/pi/releases/download/$tag/$FILENAME"
		curl -# -L -o /dev/null $DOWNLOAD_URL
		if [ $? -eq 0 ];then
			END_DOWNLOAD=`date +"%s"`
			echo "$FILENAME upload OK ($(($END_UPLOAD - $START)) seconds), download OK($(($END_DOWNLOAD-END_UPLOAD))"
		else
			echo "$FILENAME upload OK ($(($END_UPLOAD - $START)) seconds), but download failed"
		fi
	else
		echo "$FILENAME upload OK ($(($END_UPLOAD - $START)) seconds)"
	fi
else
	echo "$FILENAME upload failed"
fi



