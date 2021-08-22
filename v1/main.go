package main

import (
	"context"

	"github.com/djeeno/bqschema-gen-go/v1/generator"
)

func main() {
	ctx := context.Background()
	generator.Generate(ctx, "", "")
}
