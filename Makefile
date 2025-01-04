GO_MODULE_NAME = github.com/webbgeorge/castkeeper
FLAGS = $(shell echo "-X '$(GO_MODULE_NAME).Version=$$(git rev-parse --short HEAD)'")

install:
	npm --prefix ./web install ./web
	go mod download
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest
	$(MAKE) pre_build

pre_build:
	npm --prefix ./web run buildcss
	npm --prefix ./web run buildjs
	templ generate

run:
	$(MAKE) pre_build
	go run -ldflags="$(FLAGS)" cmd/server/main.go

watch:
	air -c cmd/server/air.toml

build:
	$(MAKE) pre_build
	go build -ldflags="$(FLAGS)" cmd/server/main.go

test:
	go vet ./...
	go test -race ./...
