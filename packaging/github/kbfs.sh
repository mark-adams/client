#!/bin/bash

# This creates a KBFS (beta) release on github from the current source/tagged version.

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

build_dir="/tmp/build_kbfs"
client_dir="$GOPATH/src/github.com/keybase/client"

echo "Loading release tool"
"$client_dir/packaging/goinstall.sh" "github.com/keybase/release"
release_bin="$GOPATH/bin/release"

version="${VERSION:-}"
if [ "$version" = "" ]; then
  version=`$release_bin latest-version --user=keybase --repo=kbfs-beta`
fi
tag="v$version"
tgz="kbfs-$version.tgz"
token="${GITHUB_TOKEN:-}"

if [ "$token" = "" ]; then
  echo "No GITHUB_TOKEN set. See https://help.github.com/articles/creating-an-access-token-for-command-line-use/"
  exit 2
fi

check_release() {
  echo "Checking for existing release: $version"
  api_url=`$release_bin url --user=keybase --repo=kbfs-beta --version=$version`
  if [ ! "$api_url" = "" ]; then
    echo "Release already exists"
    exit 0
  fi
}

build() {
  rm -rf "$build_dir"
  mkdir -p "$build_dir"
  cd "$build_dir"

  echo "Downloading source archive"
  src_url="https://github.com/keybase/kbfs-beta/archive/v$version.tar.gz"
  curl -O -J -L $src_url

  src_tgz="kbfs-beta-$version.tar.gz"
  echo "Unpacking $src_tgz"
  tar zxpf "$src_tgz"
  rm "$src_tgz"

  go_dir=/tmp/go
  rm -rf "$go_dir"
  mkdir -p "$go_dir/src/github.com"
  mv "kbfs-beta-$version" "$go_dir/src/github.com/keybase"

  echo "Building kbfs"
  GO15VENDOREXPERIMENT=1 GOPATH=$go_dir go build -a -tags "production" -o kbfs github.com/keybase/kbfs/kbfsfuse

  echo "Packaging"
  rm -rf "$tgz"
  tar zcpf "$tgz" kbfs
}

create_release() {
  cd "$build_dir"
  platform=`$release_bin platform`
  echo "Creating release"
  $release_bin create --version="$version" --repo="kbfs-beta"
  echo "Uploading release"
  $release_bin upload --src="$tgz" --dest="kbfs-$version-$platform.tgz" --version="$version" --repo="kbfs-beta"
}

check_release
build
create_release
