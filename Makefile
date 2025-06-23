GATEWAY_BIN := resolver-gateway
FULL_BIN := resolver-full
BUILD_CONF := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null)
DEBUG := DEV=true

.PHONY: build run clean

clean:
	rm ${GATEWAY_BIN} ${FULL_BIN}

build:
	${BUILD_CONF} go build -ldflags="-X main.build=${BUILD_COMMIT} -s -w" -o build/${GATEWAY_BIN} cmd/gateway/main.go
	${BUILD_CONF} go build -ldflags="-X main.build=${BUILD_COMMIT} -s -w" -o build/${FULL_BIN} cmd/full/main.go

run-gateway:
	${BUILD_CONF} ${DEBUG} go run cmd/gateway/main.go

run-full:
	${BUILD_CONF} ${DEBUG} go run cmd/full/main.go