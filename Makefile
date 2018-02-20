build:
	go build
	go vet
	golint

dep:
	dep ensure

install:
	go install

clean:
	go clean

fmt:
	gofmt -w *.go
