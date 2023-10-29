package main

import (
	"flag"
	"log"
	"os"

	"github.com/antonmisa/cliurlfetcher/internal/app"
	"github.com/antonmisa/cliurlfetcher/internal/config"
)

func main() {
	var prepare bool
	flag.BoolVar(&prepare, "prepare", false, "creating default environment and config")

	var filePath string
	flag.StringVar(&filePath, "filepath", "", "path to file with urls")

	flag.Parse()

	// Just prepare env, config and exit
	if prepare {
		err := config.Prepare()
		if err != nil {
			log.Fatalf("Prepare error: %s", err)
		}

		os.Exit(0)
	}

	// Configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg, filePath)
}
