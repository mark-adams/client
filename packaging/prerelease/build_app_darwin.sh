#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

GOPATH=${GOPATH:-}
NOPULL=${NOPULL:-} # Don't check and pull repos

if [ "$GOPATH" = "" ]; then
  echo "No GOPATH"
  exit 1
fi

build_dir_keybase="/tmp/build_keybase"
build_dir_kbfs="/tmp/build_kbfs"
client_dir="$GOPATH/src/github.com/keybase/client"
bucket_name="prerelease.keybase.io"

"$client_dir/packaging/slack/send.sh" "Starting build"

if [ ! "$NOPULL" = "1" ]; then
  "$client_dir/packaging/check_status_and_pull.sh" "$client_dir"
  "$client_dir/packaging/check_status_and_pull.sh" "$GOPATH/src/github.com/keybase/kbfs"
 else
  # Save to alternate testing bucket if we are building local
  bucket_name="prerelease-testing"
fi

BUILD_DIR=$build_dir_keybase ./build_keybase.sh
BUILD_DIR=$build_dir_kbfs ./build_kbfs.sh

cd $dir/../desktop
save_dir="/tmp/build_desktop"
rm -rf $save_dir
SAVE_DIR=$save_dir KEYBASE_BINPATH="$build_dir_keybase/keybase" KBFS_BINPATH="$build_dir_kbfs/kbfs" BUCKET_NAME=$bucket_name ./package_darwin.sh

cd $dir
SAVE_DIR=$save_dir BUCKET_NAME=$bucket_name ./s3_index.sh

"$client_dir/packaging/slack/send.sh" "Finished build. See https://s3.amazonaws.com/$bucket_name/index.html"
