install:
    go install -tags=development

install-release:
    go install

build:
    go build -tags=development

build-release:
    go build

test:
    go test ./...

vet:
    go vet ./...

check:
    go build ./...
    go test ./...
    go vet ./...
