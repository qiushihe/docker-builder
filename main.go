package main

import (
	"os"

	"github.com/qiushihe/docker-builder/builder"
)

func main() {
	os.Exit(builder.Start(os.Args))
}
