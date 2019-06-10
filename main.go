package main

import (
	"docker-builder/builder"
	"os"
)

func main() {
	os.Exit(builder.Start(os.Args))
}
