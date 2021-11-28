GO_PACKAGES := $(shell go list ./...)

build:
	go generate $(GO_PACKAGES)
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o ./delete-crashing-pods -v ./cmd/delete-crashing-pods

clean:
	go clean $(GO_PACKAGES)
	rm -fv delete-crashing-pods

test:
	golint ./...
	# go test ./...
