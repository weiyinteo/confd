#!/bin/bash

#docker run -ti --name confd-etcd3 --rm -v "$PWD":/go/src/github.com/kelseyhightower/confd -w /go/src/github.com/kelseyhightower/confd golang:1.6 bash
docker run -ti --name confd-etcd3 --rm -v "$PWD/../../..":/go/src -w /go/src/github.com/kelseyhightower/confd alpine:3.5 sh
