package main

import (
	"context"
	"flag"

	"github.com/opensibyl/gptest"
)

func main() {
	config := gptest.DefaultConfig()
	token := flag.String("token", config.Token, "openai token")
	srcPath := flag.String("src", config.SrcDir, "src")
	outputPath := flag.String("output", config.OutputDir, "output")

	// range
	before := flag.String("before", config.Before, "before")
	after := flag.String("after", config.After, "after")
	fileInclude := flag.String("include", config.FileInclude, "file include regex")
	// communication
	promptFile := flag.String("promptFile", config.After, "promptFile file")

	flag.Parse()

	if *token == "" {
		panic("token is empty")
	}

	config.Token = *token
	config.SrcDir = *srcPath
	config.OutputDir = *outputPath
	config.Before = *before
	config.After = *after
	config.FileInclude = *fileInclude
	config.PromptFile = *promptFile

	err := gptest.Run(config, context.Background())
	if err != nil {
		panic(err)
	}
}
