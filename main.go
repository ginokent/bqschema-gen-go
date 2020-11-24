//go:generate go run github.com/djeeno/bqtableschema

package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	if err := Run(ctx); err != nil {
		log.Fatalf("Run: %v\n", err)
	}
}
