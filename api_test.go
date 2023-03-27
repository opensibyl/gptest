package gptest

import (
	"context"
	"testing"
)

func TestApi(t *testing.T) {
	err := Run("", ".", context.Background())
	if err != nil {
		panic(err)
	}
}
