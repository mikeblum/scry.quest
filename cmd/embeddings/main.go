package main //nolint:revive // package comment not needed for main

import (
	"context"
	"log"
	"os"
)

func main() {
	app := NewEmbeddingsCLI()

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
