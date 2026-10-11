package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/arondubovszky/what/form"
)

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	var n = 5

	last_n_commits, err := LastNCommits(n)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var last_n_commits_str = ""
	for _, commit := range last_n_commits {
		last_n_commits_str += commit.str()
	}

	questions, err := Generate(ctx, client, last_n_commits_str, n)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	err = form.Form(questions, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
