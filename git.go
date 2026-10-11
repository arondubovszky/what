package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Commit struct {
	Hash    string
	Subject string
	Diff    string
}

func (c *Commit) str() string {
	return fmt.Sprintf("[Subject] : %s\n[Diff] : %s", c.Subject, c.Diff)
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %v: %s", args, ee.Stderr)
		}
		return "", err
	}

	return string(out), nil
}

func LastNCommits(n int) ([]Commit, error) {
	out, err := git("rev-list", "-n", strconv.Itoa(n), "HEAD") // gets the commit hashes
	if err != nil {
		return nil, err
	}

	var commits []Commit
	for _, h := range strings.Fields(out) {
		subj, err := git("show", "-s", "--format=%s", h)
		if err != nil {
			return nil, err
		}
		diff, err := git("show", "--format=", "--patch", "--first-parent", h)
		if err != nil {
			return nil, err
		}
		commits = append(commits, Commit{Hash: h, Subject: strings.TrimSpace(subj), Diff: diff})
	}

	return commits, nil
}
