package main

import (
	"github.com/rabingaire/html-parser/api"
)

func main() {
	app := api.Setup()
	app.Run(":8000")
}
