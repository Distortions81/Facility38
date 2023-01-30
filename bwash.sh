rm html/main.wasm.gz
rm html/main.wasm

GOGC=10 GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o html/main.wasm
gzip -9 html/main.wasm