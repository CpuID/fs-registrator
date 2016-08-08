package main

import (
	"log"
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
	// TODO: implement
}

func TestSyncRegistrations(t *testing.T) {
	test_esl_client := getTestEslClient(t)
	checkSipPortIsAvailable(t)
	simulateSipRegister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), "1001", "1234", uint(49202), t)
	test_kv_backend := getTestKvBackend(t)
	var test_wg sync.WaitGroup
	test_wg.Add(1)
	syncRegistrations(test_esl_client, []string{"internal"}, "192.168.99.100", 5061, 300, test_kv_backend, &test_wg, true)
	// TODO: get syncRegistrations() doing its job correctly, not populating data yet.
	result, err := test_kv_backend.Read("", true)
	// TODO: check for errors once syncRegistrations() is doing its job
	//if err != nil {
	//	t.Fatal(err)
	//}
	log.Printf("TestSyncRegistrations() Read Error: %+v\n", err)
	log.Printf("TestSyncRegistrations() Read Result: %+v\n", result)
	// TODO: check an expected_result of the above etcd read
	// Cleanup so other tests can make registrations if required.
	simulateSipDeregister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), "1001", "1234", uint(49202), t)
}
