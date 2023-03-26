package main

import (
	"context"
	"flag"

	"github.com/opensibyl/gptest"
)

func main() {
	token := flag.String("token", "", "openai token")
	flag.Parse()

	if *token == "" {
		panic("token is empty")
	}

	err := gptest.Run(*token, context.Background())
	if err != nil {
		panic(err)
	}
}
