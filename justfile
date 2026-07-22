install:
    go install

build:
    go build

test:
    go test ./...

fuzz:
    go test ./cmd -run '^$' -fuzz '^FuzzRootAcceptsArbitraryInput$'

vet:
    go vet ./...

check:
    go build ./...
    go test ./...
    go vet ./...
