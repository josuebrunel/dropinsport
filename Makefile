NAME=sdi

deps:
	go install github.com/a-h/templ/cmd/templ@latest
	go get github.com/a-h/templ@latest

test:
	go test -v -failfast -count=1 -cover -covermode=count -coverprofile=coverage.out ./...
	go tool cover -func coverage.out

templ:
	templ generate

build: templ
	go build -o bin/${NAME} cmd/main.go

run: build
	./bin/${NAME}
