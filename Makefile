# go-standards checks itself with the files it distributes.
STANDARDS_DIR := $(CURDIR)
include standard.mk

.PHONY: standard-self
standard-self:
	cp golangci.yml .golangci.yml
