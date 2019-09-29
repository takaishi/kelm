
fmt:
	go fmt ./...

run:
	go run ./main.go

build: fmt
	go build -o ik ./main.go