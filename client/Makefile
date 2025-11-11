SHELL := /bin/bash

NPM ?= npm

.PHONY: install run build clean

install:
	@if [ -f package-lock.json ]; then $(NPM) ci; else $(NPM) install; fi

run: install
	$(NPM) run dev

build: install
	$(NPM) run build

clean:
	rm -rf node_modules .next dist build
