package main

import (
	"fmt"
)

func main() {

	p := Pipeline{
		URL:     "https://github.com/sandrolain/go-ci-example.git",
		RefType: RefTypeBranch,
		Ref:     "main",
	}

	err := p.SetupWorkpath()
	if err != nil {
		panic(err)
	}

	err = p.Clone()
	if err != nil {
		panic(err)
	}

	ci, err := p.GetCI()
	if err != nil {
		panic(err)
	}

	c := Container{
		Image:    ci.Image,
		Tag:      ci.Tag,
		Workpath: p.GetWorkpath(),
		Workdir:  ci.GetWorkdir(),
		Volumes:  ci.Volumes,
	}

	err = c.RequireContainer(func(c *Container) error {
		for i, s := range ci.Steps {
			err := s.Run(c)
			if err != nil {
				return fmt.Errorf("error execution step %d: %w", i, err)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

}
