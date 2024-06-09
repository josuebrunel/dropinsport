NAME=sdi
MAIN=cmd/main.go
BIN=bin/${NAME}
MIGRATION_DIR=pkg/db/migrations
SQLBOILERFILE=sqlboiler.local.toml

deps:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	go get -u -t github.com/volatiletech/sqlboiler/v4
	go get github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql
	go install github.com/a-h/templ/cmd/templ@latest
	go get github.com/a-h/templ@latest
	go get github.com/go-playground/validator/v10

dev.deps: deps
	go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
	dev.build: templ
	go build -gcflags "all=-N -l" -o ${BIN} ${MAIN}

goose.env:
	export GOOSE_DRIVER=postgres
	export GOOSE_MIGRATION_DIR=${MIGRATION_DIR}
	goose.status: goose.env
	goose status
goose.up:
	goose up
goose.down:
	goose down

sqlgen:
	sqlboiler -c ${SQLBOILERFILE} psql

test:
	go test -v -failfast -count=1 -cover -covermode=count -coverprofile=coverage.out ./...
	go tool cover -func coverage.out

templ:
	templ generate

debug: build
	dlv --listen=:8080 --headless=true --log=true --accept-multiclient --api-version=2 exec ${BIN}

build: templ
	go build -o ${BIN} ${MAIN}

migrate.up: build
	./${BIN} migrate up
migrate.down: build
	./${BIN} migrate down

run: build
	./${BIN} --dev serve --http="0.0.0.0:8080"
	# ./bin/${NAME}
