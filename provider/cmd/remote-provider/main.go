package main

import (
	"fmt"
	"os"

	"github.com/jmorganca/ollama/api"

	ollamaprovider "github.com/iceber/wasmcloud-ollama-provider"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init ollama client: %v", err)
		os.Exit(1)
	}
	provider, err := ollamaprovider.New(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init ollama provider: %v", err)
		os.Exit(1)
	}

	if err := provider.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start ollama provider: %v", err)
		os.Exit(1)
	}
}
