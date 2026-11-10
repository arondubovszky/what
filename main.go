package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/arondubovszky/what/form"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	client := anthropic.NewClient()
	ctx := context.Background()

	setN := flag.Int("n", 3, "set the number of questions")
	setC := flag.Int("c", 3, "set the number of commits")
	flag.Parse()

	var c = *setC
	var n = *setN

	if n < 1 {
		fmt.Println("n has to be greater than 0")
		os.Exit(1)
	} else if n >= 100 {
		fmt.Println("n has to be less than a 100")
		os.Exit(1)
	}

	if c < 1 {
		fmt.Println("c has to be greater than 0")
		os.Exit(1)
	} else if c >= 100 {
		fmt.Println("c has to be less than a 100")
		os.Exit(1)
	}

	last_n_commits, err := LastNCommits(c)
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
