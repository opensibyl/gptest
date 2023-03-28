package gptest

import (
	"context"
	"testing"
)

func TestApi(t *testing.T) {
	config := DefaultConfig()
	config.Token = ""

	err := Run(config, context.Background())
	if err != nil {
		panic(err)
	}
}
