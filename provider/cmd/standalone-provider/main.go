package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/cmd"
	"github.com/spf13/cobra"

	ollamaprovider "github.com/iceber/wasmcloud-ollama-provider"
)

type ProviderConfig struct {
	WorkPath string `json:"work_path"`
}

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

	command := cmd.NewCLI()
	command.Run = nil
	command.RunE = func(_ *cobra.Command, args []string) error {
		provider, err := ollamaprovider.New(client)
		if err != nil {
			return fmt.Errorf("failed to init ollama provider: %v", err)
		}

		var config ProviderConfig
		if err := json.Unmarshal([]byte(provider.HostData().ConfigJson), &config); err != nil {
			return err
		}
		if config.WorkPath == "" {
			return errors.New("work path is not set, please specify the provider config when starting the provider")
		}

		// wasmCloud cleans up most of the environment variables when the provider is running.
		// sets the `$HOME` environment variable if it can't get the home path.
		if _, err := os.UserHomeDir(); err != nil {
			// TODO(Iceber): Windows and plan5 use different environment variables and paths.
			if err := os.Setenv("HOME", config.WorkPath); err != nil {
				return err
			}
		}

		go func() {
			if err := cmd.RunServer(nil, nil); err != nil {
				fmt.Fprintf(os.Stderr, "failed to start ollama server: %v", err)
				os.Exit(1)
			}
		}()
		if err := provider.Start(); err != nil {
			return fmt.Errorf("failed to start ollama provider: %v", err)
		}
		return nil
	}
	cobra.CheckErr(command.ExecuteContext(context.Background()))
}
