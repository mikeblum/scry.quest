package main //nolint:revive // package comment not needed for main

import (
	"context"
	"log"
	"os"

	"github.com/mikeblum/scry.quest/embeddings"
)

func main() {
	engine := embeddings.NewEngine()

	if err := engine.RunContext(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
