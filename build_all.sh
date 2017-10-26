#!/bin/bash

if ! test -d ".git"; then
  echo "needs to be in root directory" >&2
  exit 1
fi

tmp=$(mktemp -d)
version=$(cat version.go |grep Version |sed 's,.\+ "\([[:alnum:].]\+\)",\1,')

for os in "linux" "darwin" "windows"; do
  for arch in "amd64" "386" "arm"; do
    case "$arch" in
    "amd64")
      post=""
      ;;
    "386")
      post="-x86"
      ;;
    "arm")
      [ "$os" != "linux" ] && continue
      post="-arm"
      ;;
    esac

    bin="goregen"
    [ "$os" = "windows" ] && bin=$bin.exe

    target="$tmp/goregen"
    mkdir -p "$target"
    echo "building $os-$arch" >&2
    GOARCH=$arch GOOS=$os go build -o $target/$bin
    cp firmware/firmware.ino $target/
    cp default.toml $target/config.toml
    cp README.md $target/
    cd $tmp

    zipname="goregen_${version}_$os$post.zip"
    echo "bundling builds/$zipname" >&2
    zip -r $zipname goregen/
    cd -
    rm -f builds/$zipname
    mv $tmp/$zipname builds/
    rm -rf $target
  done
done

rm -rf $tmp
