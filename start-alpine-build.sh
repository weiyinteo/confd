#!/bin/bash

docker run -ti --name confd-etcd3 --rm -v "$PWD/../../..":/go/src -w /go/src/github.com/kelseyhightower/confd golang:1.7-alpine3.5 go build -v
