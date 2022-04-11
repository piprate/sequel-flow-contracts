GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all test help

all: test

test:
	(cd lib/go/iinft; make test)

help:
	@echo ''
	@echo ' Targets:'
	@echo '--------------------------------------------------'
	@echo ' all              - Run everything                '
	@echo ' test             - Run Go library unit tests     '
	@echo '--------------------------------------------------'
	@echo ''
