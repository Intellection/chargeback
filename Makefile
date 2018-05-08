SHELL := /bin/bash

install:
  glide install --strip-vendor

build:
  go build

test:
  go test -v ./...
