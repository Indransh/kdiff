default: build

build:
	@go build -ldflags "-w" -a -o bin/kdiff main.go
