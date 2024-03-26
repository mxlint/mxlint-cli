package main

import (
	"flag"

	"github.com/cinaq/mendix-model-exporter/lint"
	"github.com/sirupsen/logrus"
)

func main() {
	policies := flag.String("policies", "policies", "Path to the policies directory. Default to 'policies'")
	modelsource := flag.String("modelsource", "modelsource", "Path to the modelsource directory. Default to 'modelsource'")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	log := logrus.New()
	if *debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	lint.SetLogger(log)
	lint.EvalAll(*policies, *modelsource)
}
