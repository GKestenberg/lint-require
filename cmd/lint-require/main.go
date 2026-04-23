package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	lintrequire "github.com/giladkestenberg/lint-require"
)

func main() {
	singlechecker.Main(lintrequire.Analyzer)
}
