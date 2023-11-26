#!/bin/sh

# This is a simple helper to build the app.
# Run it from its directory such as:
#
#   ./build.sh
#

[ -d assets ] || (echo 'assets subdir is not found'; exit 1)

GOOS=js GOARCH=wasm go build -o assets/main.wasm cmd/wasm/main.go
cp -v -u "$(go env GOROOT)/misc/wasm/wasm_exec.js" assets/
