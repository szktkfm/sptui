.PHONY: all
all: build

.PHONY: build
build:
	go build -o test ./cmd/sptui

install:
	go install ./cmd/sptui

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f test
	rm ~/.config/sptui/spotify_token.json