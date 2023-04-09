package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
)

func main() {

	p := Pipeline{
		URL:     "https://github.com/sandrolain/sdt.git",
		RefType: RefTypeBranch,
		Ref:     "main",
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	temp := fmt.Sprintf("%s/temp", cwd)

	t := time.Now()
	ts := t.Format("20060102150405")

	dest := fmt.Sprintf("%s/p_%s", temp, ts)

	os.MkdirAll(dest, 0700)
	defer os.RemoveAll(dest)

	err = p.Clone(dest)
	if err != nil {
		panic(err)
	}

	c := Container{
		Image:      "golang",
		Tag:        "1.20.3",
		WorkingDir: "/workdir",
		Volumes: map[string]string{
			dest: "/workdir",
		},
	}

	commands := [][]string{
		{"ls"},
		{"sh", "./test.sh"},
	}

	err = c.RequireContainer(func(r *dockertest.Resource) error {
		for i, cmd := range commands {
			code, err := r.Exec(cmd, dockertest.ExecOptions{
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
	})
	if err != nil {
		panic(err)
	}

}
