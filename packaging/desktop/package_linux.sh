#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

client_dir="$dir/../.."
bucket_name=${BUCKET_NAME:-}
s3host="https://s3.amazonaws.com/$bucket_name"
save_dir="/tmp/build_linux"
run_mode="prod"
platform="linux"
keybase_binpath=${KEYBASE_BINPATH:-}
kbfs_binpath=${KBFS_BINPATH:-}
keybase_version=`$keybase_binpath version -S`

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

#
# TODO: Build electron for linux here
#

# App version is same as keybase service version for prerelease
app_version=$keybase_version

# Create update json for linux (no asset, only version)
mkdir -p $save_dir
$release_bin update-json --version=$app_version > $save_dir/update-$platform-$run_mode.json

# Sync to S3
s3cmd sync --acl-public --disable-multipart $save_dir/* s3://$bucket_name/
