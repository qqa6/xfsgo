
.PHONY: all
all: build

.PHONY: build
build: xfsgo

.PHONY: xfsgo
xfsgo:
	go build -o xfsgo cmd/xfsgo/main.go

.PHONY: clean
clean:
	rm -f xfsgo
