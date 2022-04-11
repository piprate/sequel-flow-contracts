GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all test help

all: test

test:
	(cd lib/go/iinft; make test)

install-gotestsum:
	 go get gotest.tools/gotestsum

test-report: install-gotestsum
	(cd lib/go/iinft; gotestsum -f testname --no-color --hide-summary failed --junitfile test-result.xml)

help:
	@echo ''
	@echo ' Targets:'
	@echo '--------------------------------------------------'
	@echo ' all              - Run everything                '
	@echo ' test             - Run Go library unit tests     '
	@echo '--------------------------------------------------'
	@echo ''
