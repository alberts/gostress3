#!/bin/bash
GOGCTRACE=1 GOMAXPROCS=12 ./6.out -test.v -test.bench=. $*
