SHELL := /bin/bash

.PHONY: dependencies build install test

dependencies:
	glide install --strip-vendor

build:
	@mkdir -p bin/
	go build -o ./bin/chargeback

install:
	go install

test:
	go test -v ./...
