#!/bin/bash
set -xe
gofmt -s -w *.go
go run gen.go
gofmt -s -w *.go
go tool 6g -I pkg main.go
go tool 6l -L pkg main.6
