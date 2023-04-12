package main

import (
	"fmt"
	"log"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	DEFAULT_TIMEOUT = 300
)

type Container struct {
	ID       string
	Image    string
	Tag      string
	Env      map[string]string
	Volumes  map[string]string
	Workpath string
	Workdir  string
	Retry    func() error
	Resource *dockertest.Resource
	Timeout  uint
}

func (c *Container) RequireContainer(exe func(*Container) error) (err error) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		err = fmt.Errorf("could not construct pool: %w", err)
		return
	}

	id := fmt.Sprintf("%s_%s", "go-ci", c.ID)

	// Create a network for our containers
	network, err := pool.CreateNetwork(id)
	if err != nil {
		log.Printf(`could not connect to docker: %s`, err)
		return
	}

	// uses pool to try to connect to Docker
	if err = pool.Client.Ping(); err != nil {
		err = fmt.Errorf("could not connect to Docker: %w", err)
		return
	}

	mounts := make([]string, len(c.Volumes)+2)
	// Mount Docker sock
	mounts[0] = "/var/run/docker.sock:/var/run/docker.sock"
	// Mount Workdir
	mounts[1] = fmt.Sprintf("%s:%s", c.Workpath, c.Workdir)

	i := 2
	for k, v := range c.Volumes {
		mounts[i] = fmt.Sprintf("%s:%s", k, v)
		i++
	}

	extraHosts := []string{"host.docker.internal:host-gateway"}

	env := make([]string, len(c.Env)+2)
	env[0] = "DOCKER_HOSTNAME=host.docker.internal"
	env[1] = "NETWORK_ID=" + network.Network.ID
	i = 2
	for k, v := range c.Env {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}

	fmt.Printf("Image %s:%s\n", c.Image, c.Tag)
	fmt.Printf("Env vars: %+v\n", env)
	fmt.Printf("Volume mounts: %+v\n", mounts)
	fmt.Printf("Workdir: %s\n", c.Workdir)
	fmt.Printf("Extra Hosts: %s\n", extraHosts)
	fmt.Printf("Network Name: %s\n", network.Network.Name)

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       id,
		Repository: c.Image,
		Tag:        c.Tag,
		Env:        env,
		WorkingDir: c.Workdir,
		Mounts:     mounts,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
		ExtraHosts: extraHosts,
		NetworkID:  network.Network.ID,
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		err = fmt.Errorf("could not start resource: %s", err)
		return
	}

	c.Resource = resource

	timeout := c.Timeout
	if timeout == 0 {
		timeout = DEFAULT_TIMEOUT
	}

	resource.Expire(timeout) // Tell docker to hard kill the container in N seconds

	// Purge removes a container and linked volumes from docker.
	defer func() {
		fmt.Printf("Purge resource\n")
		pool.Purge(resource)
	}()

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if c.Retry != nil {
		if err = pool.Retry(c.Retry); err != nil {
			err = fmt.Errorf("could not connect to container: %s", err)
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
