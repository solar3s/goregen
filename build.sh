#!/bin/bash

tmp=$(mktemp -d)

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
    echo "building $os-$arch"
    GOARCH=$arch GOOS=$os go build -o $target/$bin
    cp firmware/firmware.ino $target/
    cp README.md $target/
    cd $tmp

    zipname="goregen-$os$post.zip"
    zip -r $zipname goregen/
    cd -
    rm -f builds/$zipname
    mv $tmp/$zipname builds/
    rm -rf $target
  done
done

rm -rf $tmp
