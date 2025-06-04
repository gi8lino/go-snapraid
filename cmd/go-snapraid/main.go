package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gi8lino/go-snapraid/internal/app"
)

var (
	Version string = "dev"
	Commit  string = "none"
)

// main sets up the application context and runs the proxy.
func main() {
	ctx := context.Background()

	if err := app.Run(ctx, Version, Commit, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
