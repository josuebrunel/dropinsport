NAME=sdi

build:
	go build -o bin/${NAME} cmd/main.go

run: build
	./bin/${NAME}
