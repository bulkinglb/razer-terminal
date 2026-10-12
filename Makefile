VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
PLATFORMS := linux/amd64 linux/arm64

.PHONY: build test install release publish clean

build:
	go build -ldflags "$(LDFLAGS)"

test:
	go vet ./...
	go test ./...

install: build
	install -Dm755 razer-terminal $(HOME)/.local/bin/razer-terminal

release: test
	rm -rf dist
	mkdir dist
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; name=razer-terminal-$$os-$$arch; \
		echo "building $$name"; \
		mkdir dist/$$name && cp README.md LICENSE dist/$$name/ && \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$$name/razer-terminal . && \
		tar -C dist -czf dist/$$name.tar.gz $$name && rm -r dist/$$name || exit 1; \
	done
	cd dist && sha256sum * > checksums.txt

publish:
	@case "$(VERSION)" in v[0-9]*) ;; *) echo "set a version: make publish VERSION=v0.1.0"; exit 1;; esac
	@git diff --quiet HEAD || { echo "commit your changes first"; exit 1; }
	$(MAKE) release VERSION=$(VERSION)
	git tag $(VERSION)
	git push origin $(VERSION)
	gh release create $(VERSION) dist/* --verify-tag --title $(VERSION) --generate-notes

clean:
	rm -rf dist razer-terminal
