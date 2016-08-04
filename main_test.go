package main

import (
	"log"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
)

func teardownMain(project project.APIProject) {
	options := options.Down{
		RemoveVolume:  true,
		RemoveImages:  "local",
		RemoveOrphans: false,
	}
	err := project.Down(context.Background(), options)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	project, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFiles: []string{"docker-compose.yml"},
			ProjectName:  "fs-registrator",
		},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	ps, err := project.Ps(context.Background(), false)
	if err != nil {
		log.Fatal(err)
	}
	if len(ps) > 0 {
		// If there are any stale containers running, do a teardown first.
		teardownMain(project)
	}

	err = project.Up(context.Background(), options.Up{})
	if err != nil {
		log.Fatal(err)
	}

	exitcode := m.Run()

	teardownMain(project)

	os.Exit(exitcode)
}
