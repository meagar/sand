.PHONY: html
html:
	GOOS=js GOARCH=wasm go build -ldflags "-w" -o docs/sand.wasm ./main.go

