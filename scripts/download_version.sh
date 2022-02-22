#!/usr/bin/env bash
#
# gh-dl-release! It works!
# 
# Adapted from
#  https://gist.github.com/maxim/6e15aa45ba010ab030c4
# 
# This script downloads an asset from latest or specific Github release of a
# private repo. Feel free to extract more of the variables into command line
# parameters.
#
# PREREQUISITES
#
# curl, jq
#
# USAGE
#
# Set all the variables inside the script, make sure you chmod +x it, then
# to download specific version:
#
#     gh-dl-release 2.1.1
#
# to download latest version:
#
#     gh-dl-release latest
#
# If your version/tag doesn't match, the script will exit with error.
source /home/ubuntu/.profile # load GITHUB_PAT
TOKEN=$GITHUB_PAT
REPO="ninja-syndicate/supremacy-gameserver"
VERSION=$1                       # tag name or the word "latest"
FILE_PREFIX="gameserver"
GITHUB="https://api.github.com"

function release_curl() {
  curl -H "Authorization: token $TOKEN" \
       -H "Accept: application/vnd.github.v3.raw" \
       $@
}

if [ -z "$VERSION" ]; then
  echo " USAGE" >&2
  echo "" >&2
  echo " Set environment var \$GITHUB_PAT" >&2
  echo "" >&2
  echo "" >&2
  echo " to download specific version:" >&2
  echo "" >&2
  echo "     gh-dl-release 2.1.1" >&2
  echo "" >&2
  echo " to download latest version:" >&2
  echo "" >&2
  echo "     gh-dl-release latest" >&2
  echo "" >&2
  echo " If your version/tag doesn't match, the script will exit with error." >&2
  exit 1
fi;

if [ "$VERSION" = "latest" ]; then
  # Github should return the latest release first.
  parser_id=".assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].id"
  asset_id=`release_curl -s $GITHUB/repos/$REPO/releases/latest | jq "$parser_id"`
  parser_name=".assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].name"
  asset_name=`release_curl -s $GITHUB/repos/$REPO/releases/latest | jq --raw-output "$parser_name"`
else
  parser_id=". | map(select(.tag_name == \"$VERSION\"))[0].assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].id"
  asset_id=`release_curl -s $GITHUB/repos/$REPO/releases | jq "$parser_id"`
  parser_name=". | map(select(.tag_name == \"$VERSION\"))[0].assets | map(select(.name | contains(\"$FILE_PREFIX\")))[0].name"
  asset_name=`release_curl -s $GITHUB/repos/$REPO/releases | jq --raw-output "$parser_name"`
fi;

if [ -z "$asset_id" ]; then
  echo "ERROR: version or asset not found for $VERSION" >&2
  exit 1
fi;

if [ -t 1 ] ; then echo "Getting $asset_name"; fi # echo only if terminal

curl -H "Authorization: token $TOKEN" \
     -H "Accept:application/octet-stream" \
     --location --remote-header-name -s -o $asset_name \
     https://api.github.com/repos/$REPO/releases/assets/$asset_id \

echo $(pwd)/$asset_name
