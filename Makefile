GO_MODULE_NAME = github.com/webbgeorge/castkeeper
FLAGS = $(shell echo "-X '$(GO_MODULE_NAME).Version=$$(git rev-parse --short HEAD)'")

install:
	npm --prefix ./web install ./web
	go mod download
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
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
	go build -o cmd/server/server -ldflags="$(FLAGS)" cmd/server/main.go

test:
	$(MAKE) pre_build
	go vet ./...
	gosec ./...
	go test -race -short ./...

test_e2e:
	$(MAKE) pre_build
	go test ./e2e/... -count=1

test_cover:
	$(MAKE) pre_build
	go test -coverpkg=./... -coverprofile=profile.cov ./... -short -count=1
	go tool cover -func profile.cov

# run locally to test with alternative drivers: postgres instead of sqlite and s3 instead of local fs
run_postgres_s3:
	docker compose up -d
	$(MAKE) pre_build
	AWS_ENDPOINT_URL=http://localhost:4566 AWS_REGION=us-east-1 AWS_ACCESS_KEY_ID=000000 AWS_SECRET_ACCESS_KEY=000000 go run -ldflags="$(FLAGS)" cmd/server/main.go ./castkeeper.alt.yml

reset_postgres:
	docker compose exec postgres psql "postgresql://localdev:localdev@127.0.0.1:5432/postgres?sslmode=disable" -c "drop database castkeeper;"
	docker compose exec postgres psql "postgresql://localdev:localdev@127.0.0.1:5432/postgres?sslmode=disable" -c "create database castkeeper;"

create_user:
	go run cmd/createuser/main.go

create_user_postgres:
	go run cmd/createuser/main.go ./castkeeper.alt.yml
