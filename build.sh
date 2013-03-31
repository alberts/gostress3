#!/bin/bash
set -xe
gofmt -s -w *.go
go run gen.go
gofmt -s -w *.go
go tool 6g -I pkg main.go
go tool 6l -L pkg -L pkg/math/_test -L pkg/sync/_test -L pkg/net/http/_test/net -L pkg/strings/_test main.6
