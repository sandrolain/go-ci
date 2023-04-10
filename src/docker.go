package main

import (
	"fmt"

	"github.com/ory/dockertest/v3"
)

type Container struct {
	Image    string
	Tag      string
	Env      map[string]string
	Volumes  map[string]string
	Workpath string
	Workdir  string
	Retry    func() error
	Resource *dockertest.Resource
}

func (c *Container) RequireContainer(exe func(*Container) error) (err error) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		err = fmt.Errorf("could not construct pool: %w", err)
		return
	}

	// uses pool to try to connect to Docker
	if err = pool.Client.Ping(); err != nil {
		err = fmt.Errorf("could not connect to Docker: %w", err)
		return
	}

	env := make([]string, len(c.Env))
	i := 0
	for k, v := range c.Env {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}

	mounts := make([]string, len(c.Volumes)+1)
	i = 0
	mounts[i] = fmt.Sprintf("%s:%s", c.Workpath, c.Workdir)
	i++
	for k, v := range c.Volumes {
		mounts[i] = fmt.Sprintf("%s:%s", k, v)
		i++
	}

	fmt.Printf("image %s:%s\n", c.Image, c.Tag)
	fmt.Printf("env vars: %+v\n", env)
	fmt.Printf("volume mounts: %+v\n", mounts)
	fmt.Printf("workdir: %s\n", c.Workdir)

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: c.Image,
		Tag:        c.Tag,
		Env:        env,
		WorkingDir: c.Workdir,
		Mounts:     mounts,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
	})
	if err != nil {
		err = fmt.Errorf("could not start resource: %s", err)
		return
	}

	// Purge removes a container and linked volumes from docker.
	defer pool.Purge(resource)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if c.Retry != nil {
		if err = pool.Retry(c.Retry); err != nil {
			err = fmt.Errorf("could not connect to database: %s", err)
			return
		}
	}

	if err = exe(c); err != nil {
		return
	}
	return
}

func (c *Container) GetResource() *dockertest.Resource {
	return c.Resource
}
