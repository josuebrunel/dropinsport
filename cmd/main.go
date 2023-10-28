package main

import (
	"os"

	"github.com/josuebrunel/sportdropin/app"
)

func main() {
	var listenAddr = ":8080"
	if v := os.Getenv("SDI_HTTP"); v != "" {
		listenAddr = v
	}
	server := app.NewApp(listenAddr)
	server.Run()
}
