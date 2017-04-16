#!/bin/bash

tmp=$(mktemp -d)

GOARCH=amd64 GOOS=linux go build -o $tmp/goregen-linux
GOARCH=386 GOOS=linux go build -o $tmp/goregen-linux-x86
GOARCH=arm GOOS=linux go build -o $tmp/goregen-linux-arm

GOARCH=amd64 GOOS=darwin go build -o $tmp/goregen-darwin
GOARCH=386 GOOS=darwin go build -o $tmp/goregen-darwin-x86

GOARCH=amd64 GOOS=windows go build -o $tmp/goregen-win.exe
GOARCH=386 GOOS=windows go build -o $tmp/goregen-win-x86.exe


for i in $(ls -1 $tmp/*); do
  tmp2=$(mktemp -d)
  target="$tmp2/goregen"
  mkdir $target
  cp $i $target/
  cp firmware/firmware.ino $target/

  cd $tmp2
  zip -r $(basename $i).zip goregen/
  cd -
  rm -f builds/$(basename $i).zip
  mv $tmp2/$(basename $i).zip builds/
  rm -rf $tmp2
done

rm -rf $tmp

