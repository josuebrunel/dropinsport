package main

import (
	"github.com/josuebrunel/sportdropin/app"
)

func main() {
	server := app.NewApp()
	server.Run()
}
