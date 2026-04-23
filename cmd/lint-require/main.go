package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	lintrequire "github.com/GKestenberg/lint-require"
)

func main() {
	singlechecker.Main(lintrequire.Analyzer)
}
