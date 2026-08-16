package main

import (
	"log"
	"os"

	"github.com/kazeyukiro/3m-ui/backend/internal/app"
)

// main accepts both the installer-provided THREE_M_UI_CONFIG environment
// variable and the documented --config/-c command-line option.
func main() {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version", "version":
			println(versionString())
			return
		case "--config", "-c":
			if i+1 >= len(args) || args[i+1] == "" {
				log.Fatal("--config requires a path")
			}
			if err := os.Setenv("THREE_M_UI_CONFIG", args[i+1]); err != nil {
				log.Fatalf("set config path: %v", err)
			}
			i++
		default:
			log.Fatalf("unknown argument: %s", args[i])
		}
	}

	if err := app.Run(frontendFiles); err != nil {
		log.Fatal(err)
	}
}
