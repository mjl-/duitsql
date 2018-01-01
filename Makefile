build:
	go vet
	golint
	go build

install:
	go install

clean:
	go clean

fmt:
	gofmt -w *.go
