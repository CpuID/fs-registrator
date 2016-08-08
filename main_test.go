package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

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

// Fetches the IP from DOCKER_HOST env var.
// If it is not set, assume its a local install and return 127.0.0.1
func getDockerHost() (string, error) {
	docker_host_env := os.Getenv("DOCKER_HOST")
	if len(docker_host_env) == 0 {
		log.Printf("Assuming local Docker installation (127.0.0.1)\n")
		return "127.0.0.1", nil
	}
	// Example: tcp://192.168.99.100:2376
	split_docker_host := strings.Split(docker_host_env, ":")
	if len(split_docker_host) != 3 {
		return "", errors.New(fmt.Sprintf("Invalid format for DOCKER_HOST environment variable, expected proto://host:port, got %s", docker_host_env))
	}
	docker_host := strings.Replace(split_docker_host[1], "/", "", 2)
	log.Printf("DOCKER_HOST environment variable set, using %s\n", docker_host)
	return docker_host, nil
}

// map["servicename-internalport/protocol"] = externalport
func parseContainerPorts(info_set *project.InfoSet, docker_project_name string) (map[string]uint, error) {
	result := make(map[string]uint)
	container_keys := make(map[int]string)
	for k1, v1 := range *info_set {
		for _, v2 := range v1 {
			//log.Printf("k1: %d, k2: %d, v2 Key: %s, v2 Val: %s\n", k1, k2, v2.Key, v2.Value)
			if v2.Key == "Name" {
				container_keys[k1] = v2.Value
			} else if v2.Key == "Ports" {
				// If we don't have the container name already, can't proceed for it.
				if _, ok := container_keys[k1]; ok == false {
					return map[string]uint{}, errors.New("parseContainerPorts() : Found 'Ports' for a Container Key without a 'Name' attribute prior, cannot proceed.")
				}
				// v2.Value == 2380/tcp, 0.0.0.0:32777->2379/tcp
				// Parse out the ports that are exposed publicly.
				// Could use regex for some of these splits, being super safe and just doing it with basic string splits.
				for _, v3 := range strings.Split(v2.Value, ", ") {
					if strings.Contains(v3, "->") == true {
						// Port is exposed outside of the container.
						split_host_port := strings.Split(v3, "->")
						if len(split_host_port) != 2 {
							return map[string]uint{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Splitting host/ports '%s' (by ->) expected 2 results, got %d, cannot proceed.",
								v3, len(split_host_port)))
						}
						split_external_host_port := strings.Split(split_host_port[0], ":")
						if len(split_external_host_port) != 2 {
							return map[string]uint{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Splitting external host/port '%s' (by :) expected 2 results, got %d, cannot proceed.",
								split_host_port[0], len(split_external_host_port)))
						}
						split_internal_port := strings.Split(split_host_port[1], "/")
						if len(split_internal_port) != 2 {
							return map[string]uint{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Splitting internal port '%s' (by /) expected 2 results, got %d, cannot proceed.",
								split_host_port[1], len(split_internal_port)))
						}
						// TODOLATER: validate the pieces of split_host and split_port?
						external_port_int, err := strconv.Atoi(split_external_host_port[1])
						if err != nil {
							return map[string]uint{}, errors.New(fmt.Sprintf(
								"parseContainerPorts() : Cannot convert %s to an int for external port, cannot proceed. Error: %s",
								split_external_host_port[1], err.Error()))
						}
						use_service_name := strings.Replace(container_keys[k1], fmt.Sprintf("%s_", docker_project_name), "", 1)
						result[fmt.Sprintf("%s-%s/%s", use_service_name, split_internal_port[0], split_internal_port[1])] = uint(external_port_int)
					}
				}
			}
		}
	}
	log.Printf("Service Ports: %+v\n", result)
	return result, nil
}

// Performs a poor mans health check on the underlying container ports, make sure the services are up before proceeding.
// TODOLATER: try use the new HEALTHCHECK attribute in Docker 1.12, would need to pull in a docker api client lib to retrieve
// inspect() most likely.
func pollContainerTcpPortHealth(host string, port uint, timeout_sec uint) error {
	log.Printf("Attempting connection to TCP %s:%d...", host, port)
	for i := uint(0); i < timeout_sec; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			fmt.Printf(".")
		} else {
			conn.Close()
			fmt.Printf(" Success.\n")
			return nil
		}
		time.Sleep(time.Second)
	}
	fmt.Printf(" Timeout reached.\n")
	return errors.New(fmt.Sprintf("Connection attempt to %s:%d timed out after %d seconds.", host, port, timeout_sec))
}

var dockerHost string
var dockerContainerPorts map[string]uint

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
	defer teardownMain(project)

	ps, err = project.Ps(context.Background(), false)
	if err != nil {
		log.Fatal(err)
	}

	docker_host, err := getDockerHost()
	if err != nil {
		log.Fatal(err)
	}
	dockerHost = docker_host

	container_ports, err := parseContainerPorts(&ps, docker_project_name)
	if err != nil {
		log.Fatal(err)
	}
	dockerContainerPorts = container_ports
	// Attempt to connect to all the ports before proceeding.
	for k, v := range dockerContainerPorts {
		split_k := strings.Split(k, "/")
		// We only health check TCP ports for now, safe enough.
		if split_k[1] == "tcp" {
			err = pollContainerTcpPortHealth(docker_host, v, uint(20))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	// Sleep for 30 seconds before proceeding, to ensure Travis tests pass.
	// Above we do health checks on the TCP ports, but for some reason
	// they pass immediately on Travis without waiting for things like event_socket
	// to start.
	// https://github.com/CpuID/fs-registrator/issues/4
	if os.Getenv("TRAVIS") == "true" {
		log.Printf("Sleeping for 30 seconds (to keep Travis CI happy)...\n")
		time.Sleep(30 * time.Second)
	} else {
		// Give SIP an extra second to start after event_socket to be safe, since we don't
		// health check on UDP ports.
		time.Sleep(time.Second)
	}

	exitcode := m.Run()

	teardownMain(project)

	os.Exit(exitcode)
}
