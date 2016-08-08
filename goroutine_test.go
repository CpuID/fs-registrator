package main

import (
	//"log"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

func getTestKvBackend(t *testing.T) KvBackend {
	if _, ok := dockerContainerPorts["etcd_1-2379/tcp"]; ok == false {
		t.Fatal("Docker Container port for etcd not found in dockerContainerPorts, did the container start?")
	}
	test_kv_backend, err := CreateKvBackend(map[string]string{
		"backend": "etcd",
		"host":    dockerHost,
		"port":    strconv.Itoa(int(dockerContainerPorts["etcd_1-2379/tcp"])),
		"prefix":  "fs_test_registrations",
	})
	if err != nil {
		t.Fatal(err)
	}
	return test_kv_backend
}

func TestWatchForRegistrationEvents(t *testing.T) {
	test_esl_client := getTestEslClient(t)
	test_kv_backend := getTestKvBackend(t)
	test_advertise_ip := "192.168.99.100"
	test_advertise_port := 5062
	test_sip_user := "1002"
	test_sip_pass := "1234"
	test_sip_contact_port := uint(49203)
	expected_result1 := map[string]string{
		"1002@sip.testserver.tld": "{\"host\":\"192.168.99.100\",\"port\":5062}",
	}
	// result 2 is an empty map

	var test_wg sync.WaitGroup
	go test_esl_client.Handle()
	// This channel is triggered on each event execution.
	event_channel := make(chan struct{})
	defer close(event_channel)

	// Start our watcher which will update the K/V store on changes.
	test_wg.Add(1)
	go watchForRegistrationEvents(test_esl_client, test_advertise_ip, test_advertise_port, test_kv_backend, &test_wg, 3, event_channel)

	// Make sure that we are watching for events before proceeding.
	//log.Printf("Wait for event: 1\n")
	<-event_channel
	//log.Printf("Done waiting for event: 1\n")

	// Do a SIP register, so we have something to start with.
	simulateSipRegister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), test_sip_user, test_sip_pass, test_sip_contact_port, t)

	// Wait until an event is processed before proceeding.
	//log.Printf("Wait for event: 2\n")
	<-event_channel
	//log.Printf("Done waiting for event: 2\n")

	// Check it's in etcd.
	result1, err := test_kv_backend.Read("", true)
	if err != nil {
		t.Fatal(err)
	}
	//log.Printf("TestWatchForRegistrationEvents() Read Error: %+v\n", err)
	//log.Printf("TestWatchForRegistrationEvents() Read Result 1: %+v\n", result1)
	if reflect.DeepEqual(*result1, expected_result1) != true {
		t.Error("Expected", expected_result1, "got", result1)
	}

	// Cleanup the registration, before performing another sync.
	simulateSipDeregister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), test_sip_user, test_sip_pass, test_sip_contact_port, t)

	// Wait until an event is processed before proceeding.
	//log.Printf("Wait for event: 3\n")
	<-event_channel
	//log.Printf("Done waiting for event: 3\n")

	// Check it's gone from etcd.
	result2, err := test_kv_backend.Read("", true)
	if err != nil {
		t.Fatal(err)
	}
	//log.Printf("TestWatchForRegistrationEvents() Read Error 2: %+v\n", err)
	//log.Printf("TestWatchForRegistrationEvents() Read Result 2: %+v\n", result2)
	if len(*result2) > 0 {
		t.Errorf("Expected a zero length result from K/V backend, got %d results: %+v\n", len(*result2), *result2)
	}
}

func TestSyncRegistrations(t *testing.T) {
	test_esl_client := getTestEslClient(t)
	test_kv_backend := getTestKvBackend(t)
	checkSipPortIsAvailable(t)
	test_sofia_profiles := []string{"internal"}
	test_advertise_ip := "192.168.99.100"
	test_advertise_port := 5061
	test_sip_user := "1001"
	test_sip_pass := "1234"
	test_sip_contact_port := uint(49202)
	expected_result1 := map[string]string{
		"1001@sip.testserver.tld": "{\"host\":\"192.168.99.100\",\"port\":5061}",
	}
	// result 2 is an empty map

	// Do a SIP register, so we have something to start with.
	simulateSipRegister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), test_sip_user, test_sip_pass, test_sip_contact_port, t)

	var test_wg sync.WaitGroup
	go test_esl_client.Handle()

	// First sync, should perform an add to the K/V backend.
	test_wg.Add(1)
	syncRegistrations(test_esl_client, test_sofia_profiles, test_advertise_ip, test_advertise_port, 300, test_kv_backend, &test_wg, true)
	result1, err := test_kv_backend.Read("", true)
	if err != nil {
		t.Fatal(err)
	}
	//log.Printf("TestSyncRegistrations() Read Error: %+v\n", err)
	//log.Printf("TestSyncRegistrations() Read Result 1: %+v\n", result1)
	if reflect.DeepEqual(*result1, expected_result1) != true {
		t.Error("Expected", expected_result1, "got", result1)
	}

	// Cleanup the registration, before performing another sync.
	simulateSipDeregister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), test_sip_user, test_sip_pass, test_sip_contact_port, t)

	// Second sync, should perform a remove from the K/V backend.
	test_wg.Add(1)
	syncRegistrations(test_esl_client, test_sofia_profiles, test_advertise_ip, test_advertise_port, 300, test_kv_backend, &test_wg, true)
	result2, err := test_kv_backend.Read("", true)
	if err != nil {
		t.Fatal(err)
	}
	//log.Printf("TestSyncRegistrations() Read Error 2: %+v\n", err)
	//log.Printf("TestSyncRegistrations() Read Result 2: %+v\n", result2)
	if len(*result2) > 0 {
		t.Errorf("Expected a zero length result from K/V backend, got %d results: %+v\n", len(*result2), *result2)
	}
}
