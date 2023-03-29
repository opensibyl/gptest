package main

import (
	"context"
	"flag"
	"os"

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
	promptFile := flag.String("promptFile", config.PromptFile, "promptFile file")

	flag.Parse()

	// trying to read token from env
	tokenFromEnv := os.Getenv("OPENAI_TOKEN")
	if tokenFromEnv != "" {
		*token = tokenFromEnv
	}

	// still empty
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
