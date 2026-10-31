package main

import (
	"flag"
	"log"
	"strings"

	"github.com/nikimonax/go-metrics/tools/structreset"
)

func main() {
	var patternsRaw string

	flag.StringVar(
		&patternsRaw,
		"packages",
		"./...",
		"list of packages for codegen",
	)
	flag.Parse()

	if patternsRaw == "" {
		log.Fatalln("required non-empty 'packages'")
	}

	packages := strings.Split(patternsRaw, ",")

	if err := structreset.Main(packages...); err != nil {
		log.Fatalln(err)
	}
}
