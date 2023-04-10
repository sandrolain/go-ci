package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-shellwords"
	"github.com/ory/dockertest/v3"
)

type Steps []Step

type Step struct {
	Commands  []string
	Artifacts []string
}

func (s *Step) Run(c *Container) error {
	r := c.GetResource()
	for i, cmd := range s.Commands {
		words, err := shellwords.Parse(cmd)
		if err != nil {
			return fmt.Errorf("cmd %v: cannot parse '%s': error: %w", i, cmd, err)
		}
		fmt.Printf("%+v\n", s)
		fmt.Printf("exec command '%s', %d words %+v\n", cmd, len(words), words)
		code, err := r.Exec(words, dockertest.ExecOptions{
			StdOut: os.Stdout,
			StdErr: os.Stderr,
		})
		if err != nil {
			return fmt.Errorf("cmd %v: error: %w", i, err)
		}
		if code > 0 {
			return fmt.Errorf("cmd %v: exit code: %d", i, code)
		}
	}
	return nil
}
