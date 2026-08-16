package main

import (
	"embed"
	"log"
	"os"

	"github.com/kazeyukiro/3m-ui/backend/internal/app"
)

//go:embed web/dist/*
var frontendFiles embed.FS

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		println(versionString())
		return
	}
	if err := app.Run(frontendFiles); err != nil {
		log.Fatal(err)
	}
}
