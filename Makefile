GOPATH := $(shell pwd)
.PHONY: clean test

all:
	@GOPATH=$(GOPATH) go install huobi 

clean:
	@rm -fr bin pkg
