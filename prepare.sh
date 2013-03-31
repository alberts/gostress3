#!/bin/bash
set -xe
export TMPDIR=`pwd`/tmp
rm -rf $TMPDIR pkg testdata
mkdir -p $TMPDIR
go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.Dir}}{{end}}' std | xargs -t -n 1 -I PKGDIR sh -c 'cd PKGDIR && go test -work -c'
mkdir -p pkg
cp -a $TMPDIR/*/* pkg
rm -rf $TMPDIR
find $GOROOT/src -name testdata -type d | xargs cp -at .
