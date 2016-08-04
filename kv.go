package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type KvBackend interface {
	Read(key string, recursive bool) (*map[string]string, error)
	Write(key string, value string, ttl int) error
	Delete(key string) error
}

// Credit: http://matthewbrown.io/2016/01/23/factory-pattern-in-golang/

func init() {
	RegisterKvBackend("etcd", NewKvBackendEtcd)
	// Add new backends here as they become available.
}

type KvBackendFactory func(conf map[string]string) (KvBackend, error)

var kvBackendFactories = make(map[string]KvBackendFactory)

func RegisterKvBackend(name string, factory KvBackendFactory) {
	if factory == nil {
		log.Fatalf("K/V backend factory '%s' does not exist.", name)
	}
	_, registered := kvBackendFactories[name]
	if registered {
		log.Printf("K/V backend factory '%s' already registered. Ignoring.", name)
	}
	kvBackendFactories[name] = factory
}

// Make a list of all available K/V backend factories
func availableKvBackends() []string {
	var available_kv_backends []string
	for k, _ := range kvBackendFactories {
		available_kv_backends = append(available_kv_backends, k)
	}
	return available_kv_backends
}

func CreateKvBackend(conf map[string]string) (KvBackend, error) {
	if _, ok := conf["backend"]; ok == false {
		return nil, errors.New("'backend' key does not exist in conf.")
	}

	kvBackendFactory, ok2 := kvBackendFactories[conf["backend"]]

	if ok2 == false {
		// Factory has not been registered
		return nil, errors.New(fmt.Sprintf("Invalid K/V Backend Name. Must be one of: %s", strings.Join(availableKvBackends(), ", ")))
	}

	// Run the factory with the configuration
	return kvBackendFactory(conf)
}

//

type KvBackendValue struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func getKvBackendValueType(ip string, port int) KvBackendValue {
	return KvBackendValue{
		Host: ip,
		Port: port,
	}
}

func getKvBackendValueJsonType(input string) (KvBackendValue, error) {
	var result KvBackendValue
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return KvBackendValue{}, err
	}
	return result, nil
}

func getKvBackendValueJsonString(input KvBackendValue) (string, error) {
	json, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(json), nil
}
