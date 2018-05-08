SHELL := /bin/bash

.PHONY: dependencies build test

dependencies:
	glide install --strip-vendor

build:
	@mkdir -p bin/
	go build -o ./bin/chargeback

test:
	go test -v ./...
