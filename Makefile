build:
	CGO_ENABLED=0 \
    GOOS=linux \
    go build -buildvcs=false -o sse
.PHONY: build

run: clean build
	./sse
.PHONY: run

clean:
	@rm sse
.PHONY: clean
