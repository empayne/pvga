#!/bin/sh
# Install dependencies for libxml2, this is a bit hacky:
apk update && apk add pkgconfig libxml2-dev gcc build-base
go run main.go # TODO: should build binaries vs. 'go run' on source files