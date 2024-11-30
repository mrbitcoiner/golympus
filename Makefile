VERSION = v0.1.0

build:
	[ -e "out" ] || mkdir out
	go build \
		-o out/golympus \
		-ldflags '-s -w -X main.version=$(VERSION)' \
		.

test:
	go test . -v -count=1 -p=1
