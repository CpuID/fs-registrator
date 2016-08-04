package main

import (
	"log"
	"testing"

	"github.com/0x19/goesl"
)

func TestSubscribeToFreeswitchRegEvents(t *testing.T) {
	if _, ok := dockerContainerPorts["freeswitch_1-8021"]; ok == false {
		t.Fatal("Docker Container port for FreeSWITCH ESL not found in dockerContainerPorts, did the container start?")
	}
	log.Printf("TestSubscribeToFreeswitchRegEvents() : Docker Container FreeSWITCH ESL Port - %d\n", uint(dockerContainerPorts["freeswitch_1-8021"]))
	// TODO: verify host for NewClient below. needs to come from libcompose input.
	test_client, err := goesl.NewClient("127.0.0.1", uint(dockerContainerPorts["freeswitch_1-8021"]), "ClueCon", int(5))
	if err != nil {
		t.Fatal(err)
	}
	err = subscribeToFreeswitchRegEvents(&test_client)
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
}

func TestParseFreeswitchRegEvent(t *testing.T) {
	// TODO: implement
}

func TestGetFreeswitchRegistrations(t *testing.T) {
	// TODO: implement
}
