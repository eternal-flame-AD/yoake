PROJECT_NAME := yoake
MODULE_PATH := github.com/eternal-flame-AD/${PROJECT_NAME}

CMD_DIR := cmd

COMMANDS := $(patsubst ${CMD_DIR}/%,%,$(shell find ${CMD_DIR}/ -mindepth 1 -maxdepth 1 -type d))
COMMANDSDIST = $(addprefix dist/,${COMMANDS})
ifeq ($(INSTALLDEST),)
INSTALLDEST := /opt/${PROJECT_NAME}
endif

install:
	mkdir -p $(INSTALLDEST)
	cp -r dist/* $(INSTALLDEST)

build: webroot $(COMMANDSDIST)
	chmod -R 755 $(COMMANDSDIST)

dev:
	while true; do \
		kill $$(cat .server.pid); \
		make build && \
			(dist/server -c config-dev.yml & echo $$! > .server.pid); \
		inotifywait -e modify -r webroot internal server config && kill $(cat .server.pid) ; \
	done

webroot: $(wildcard webroot/**) FORCE
	cp -r webroot dist
	(cd dist/webroot; ../../scripts/webroot-build.fish)

verify:
	go vet ./...
	go mod verify

clean:
	rm -rf dist/webroot
	rm -rf dist

dist/%: ${CMD_DIR}/% FORCE
	go build -o $@ ${MODULE_PATH}/$<

.PHONY: build clean
FORCE: