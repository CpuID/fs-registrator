package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"golang.org/x/net/context"
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

// map["servicename-internalport"] = externalport
func parseContainerPorts(info_set *project.InfoSet, docker_project_name string) (map[string]int, error) {
	result := make(map[string]int)
	container_keys := make(map[int]string)
	for k1, v1 := range *info_set {
		for k2, v2 := range v1 {
			//log.Printf("k1: %d, k2: %d, v2 Key: %s, v2 Val: %s\n", k1, k2, v2.Key, v2.Value)
			if v2.Key == "Name" {
				container_keys[k1] = v2.Value
			} else if v2.Key == "Ports" {
				// If we don't have the container name already, can't proceed for it.
				if _, ok := container_keys[k1]; ok == false {
					return map[string]int{}, errors.New("parseContainerPorts() : Found 'Ports' for a Container Key without a 'Name' attribute prior, cannot proceed.")
				}
				// v2.Value == 2380/tcp, 0.0.0.0:32777->2379/tcp
				// Parse out the ports that are exposed publicly.
				// Could use regex for some of these splits, being super safe and just doing it with basic string splits.
				for _, v3 := range strings.Split(v2.Value, ", ") {
					if strings.Contains(v3, "->") == true {
						// Port is exposed outside of the container.
						split_host_port := strings.Split(v3, "->")
						// We ignore protocol right now, and assume that the executor that uses this knows what protocol it will use (TCP/UDP).
						if len(split_host_port) != 2 {
							return map[string]int{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Splitting host/ports '%s' (by ->) expected 2 results, got %d, cannot proceed.",
								v3, len(split_host_port)))
						}
						split_external_host_port := strings.Split(split_host_port[0], ":")
						if len(split_external_host_port) != 2 {
							return map[string]int{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Splitting external host/port '%s' (by :) expected 2 results, got %d, cannot proceed.",
								split_host_port[0], len(split_external_host_port)))
						}
						split_internal_port := strings.Split(split_host_port[1], "/")
						if len(split_internal_port) != 2 {
							return map[string]int{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Splitting internal port '%s' (by /) expected 2 results, got %d, cannot proceed.",
								split_host_port[1], len(split_internal_port)))
						}
						// TODOLATER: validate the pieces of split_host and split_port?
						external_port_int, err := strconv.Atoi(split_external_host_port[1])
						if err != nil {
							return map[string]int{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Cannot convert %s to an int for external port, cannot proceed. Error: %s",
								split_external_host_port[1], err.Error()))
						}
						use_service_name := strings.Replace(container_keys[k1], fmt.Sprintf("%s_", docker_project_name), "", 1)
						result[fmt.Sprintf("%s-%s", use_service_name, split_internal_port[0])] = external_port_int
					}
				}
			}
		}
	}
	log.Printf("Service Ports: %+v\n", result)
	return result, nil
}

var dockerContainerPorts map[string]int

func TestMain(m *testing.M) {
	docker_project_name := "fsregistrator"
	project, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFiles: []string{"docker-compose.yml"},
			ProjectName:  docker_project_name,
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

	ps, err = project.Ps(context.Background(), false)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: we need to try use the HEALTHCHECK feature in Docker 1.12 so that we can verify FreeSWITCH is up
	// and responding to TCP 8021

	container_ports, err := parseContainerPorts(&ps, docker_project_name)
	if err != nil {
		log.Fatal(err)
	}
	dockerContainerPorts = container_ports

	exitcode := m.Run()

	teardownMain(project)

	os.Exit(exitcode)
}
