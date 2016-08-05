package main

import (
	"log"
	"reflect"
	"testing"

	"github.com/0x19/goesl"
)

func getTestEslClient(t *testing.T) *goesl.Client {
	if _, ok := dockerContainerPorts["freeswitch_1-8021"]; ok == false {
		t.Fatal("Docker Container port for FreeSWITCH ESL not found in dockerContainerPorts, did the container start?")
	}
	log.Printf("getTestEslClient() : Docker Container FreeSWITCH ESL Port - %d\n", uint(dockerContainerPorts["freeswitch_1-8021"]))
	test_client, err := goesl.NewClient(dockerHost, uint(dockerContainerPorts["freeswitch_1-8021"]), "ClueCon", int(5))
	if err != nil {
		t.Fatal(err)
	}
	return &test_client
}

func TestSubscribeToFreeswitchRegEvents(t *testing.T) {
	test_client := getTestEslClient(t)
	err := subscribeToFreeswitchRegEvents(test_client)
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
}

func TestParseFreeswitchRegEvent(t *testing.T) {
	// TODO: fill in
	expected_result1 := ""
	expected_result2 := ""
	input := goesl.Message{
		Headers: map[string]string{
		// TODO: fill in
		//"": "",
		},
		Body: []byte{},
	}
	result1, result2, err := parseFreeswitchRegEvent(&input)
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
	if result1 != expected_result1 {
		t.Error("Expected", expected_result1, "got", result1)
	}
	if result2 != expected_result2 {
		t.Error("Expected", expected_result2, "got", result2)
	}
}

func TestGetFreeswitchRegistrations(t *testing.T) {
	expected_result := []string{
	// TODO: fill in
	}
	test_client := getTestEslClient(t)
	result, err := getFreeswitchRegistrations(test_client, []string{
	// TODO: fill in
	})
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
	if reflect.DeepEqual(result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
}
