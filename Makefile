SHELL := /bin/bash

dependencies:
	glide install --strip-vendor

build:
	go build

test:
	go test -v ./...
