package main

import (
	"fmt"
	"os"

	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/cmd"

	ollamaprovider "github.com/iceber/wasmcloud-ollama-provider"
)

func main() {
	// It is inefficient to use a client to connect to the internally started server.
	// The best way to do this is to call the methods in ollama/server directly,
	// but the server package doesn't expose an elegant way to do this,
	// so for the time being, we'll use a client to connect.
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

	go func() {
		if err := cmd.RunServer(nil, nil); err != nil {
			fmt.Fprintf(os.Stderr, "failed to start ollama server: %v", err)
			os.Exit(1)
		}
	}()
	if err := provider.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start ollama provider: %v", err)
		os.Exit(1)
	}
}
