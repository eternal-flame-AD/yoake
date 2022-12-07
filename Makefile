PROJECT_NAME := yoake
MODULE_PATH := github.com/eternal-flame-AD/${PROJECT_NAME}

CMD_DIR := cmd
WASM_DIR := wasm

COMMANDS := $(patsubst ${CMD_DIR}/%,%,$(shell find ${CMD_DIR}/ -mindepth 1 -maxdepth 1 -type d))
WASM_APPS := $(patsubst ${WASM_DIR}/%,%.wasm,$(shell find ${WASM_DIR}/ -mindepth 1 -maxdepth 1 -type d))

COMMANDSDIST = $(addprefix dist/,${COMMANDS})
WASM_APPSDIST = $(addprefix dist/web/,${WASM_APPS})
ifeq ($(INSTALLDEST),)
INSTALLDEST := /opt/${PROJECT_NAME}
endif

VERSION := $(shell git describe --tags --exact HEAD || printf "%s" $(shell git rev-parse --short HEAD))
BUILDDATE := $(shell date -Iminutes)

install:
	mkdir -p $(INSTALLDEST)
	cp -r dist/* $(INSTALLDEST)

build: webroot $(COMMANDSDIST) $(WASM_APPSDIST)
	chmod -R 755 $(COMMANDSDIST) $(WASM_APPSDIST)

dev:
	while true; do \
		kill $$(cat .server.pid); \
		make GOGCFLAGS='all=-N -l' build && \
			(dist/server -c config-dev.yml & echo $$! > .server.pid); \
			jq " .configurations[0].processId = $$(cat .server.pid) " .vscode/launch-tpl.json > .vscode/launch.json; \
		inotifywait -e modify -r webroot internal server config && kill $(cat .server.pid) ; \
	done

webroot: $(wildcard webroot/**) FORCE
	mkdir -p dist
	mkdir -p dist/web
	cp -r assets dist
	cp -r webroot dist
	(cd dist/webroot; ../../scripts/webroot-build.fish)

verify:
	go vet ./...
	go mod verify

clean:
	rm -rf dist/webroot
	rm -rf dist

dist/web/%.wasm: ${WASM_DIR}/% FORCE
	GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -buildvcs\
		-ldflags "-X ${MODULE_PATH}/internal/version.tagVersion=$(VERSION) 	\
				  -X ${MODULE_PATH}/internal/version.buildDate=$(BUILDDATE)  \
					-s -w" \
		-o $@ ${MODULE_PATH}/$<

dist/%: ${CMD_DIR}/% FORCE
	go build -buildvcs\
		-ldflags "-X ${MODULE_PATH}/internal/version.tagVersion=$(VERSION) 	\
				  -X ${MODULE_PATH}/internal/version.buildDate=$(BUILDDATE)" \
		-gcflags "$(GOGCFLAGS)" \
		-o $@ ${MODULE_PATH}/$<


.PHONY: build clean
FORCE: