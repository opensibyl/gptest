package main

import (
	"context"
	"flag"

	"github.com/opensibyl/gptest"
)

func main() {
	token := flag.String("token", "", "openai token")
	srcPath := flag.String("src", ".", "src")
	outputPath := flag.String("output", ".", "output")
	flag.Parse()

	if *token == "" {
		panic("token is empty")
	}

	config := gptest.DefaultConfig()
	config.Token = *token
	config.SrcDir = *srcPath
	config.OutputDir = *outputPath

	err := gptest.Run(config, context.Background())
	if err != nil {
		panic(err)
	}
}
