.PHONY: all
all: build

.PHONY: build
build:
	go build -o test ./cmd/spotui

install:
	go install ./cmd/spotui

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f test
	rm ~/.config/spotui/spotify_token.json