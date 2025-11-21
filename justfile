install:
    go install -tags=development

install-release:
    go install

build:
    go build -tags=development

build-release:
    go build
