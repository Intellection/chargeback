SHELL := /bin/bash

.PHONY: dependencies build test

dependencies:
	glide install --strip-vendor

build:
	go build

test:
	go test -v ./...
