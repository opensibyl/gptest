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

	err := gptest.Run(*token, *srcPath, *outputPath, context.Background())
	if err != nil {
		panic(err)
	}
}
