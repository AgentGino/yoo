package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/agentgino/yo/internal/cli"
)

// Version is injected by release builds and Homebrew formula ldflags.
var Version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := cli.Runner{Stdout: os.Stdout, Stderr: os.Stderr, Version: Version}
	if err := runner.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "yo: %v\n", err)
		os.Exit(1)
	}
}
