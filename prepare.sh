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
find pkg -type l -name '*.a' -print0 | xargs -0 rm -f
find `pwd`/pkg -name '*.a' -a ! -name '*_test.a' -a ! -path '*/_test/main.a' | perl -ne 'chomp;$p=$_;s/.*\/_test\///;`ln -sf $p pkg/$_`;'
