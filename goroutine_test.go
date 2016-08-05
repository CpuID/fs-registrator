package main

import (
	"strconv"
	//"sync"
	"testing"
)

func getTestKvBackend(t *testing.T) KvBackend {
	if _, ok := dockerContainerPorts["etcd_1-2379"]; ok == false {
		t.Fatal("Docker Container port for etcd not found in dockerContainerPorts, did the container start?")
	}
	test_kv_backend, err := CreateKvBackend(map[string]string{
		"backend": "etcd",
		"host":    dockerHost,
		"port":    strconv.Itoa(int(dockerContainerPorts["etcd_1-2379"])),
		"prefix":  "fs_registrations",
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
	//test_esl_client := getTestEslClient(t)
	//err := simulateSipRegister(dockerHost, uint(5060), "", "", t)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//test_kv_backend := getTestKvBackend(t)
	// TODO: populate an entry for the same advertise ip/port that will get removed as part of a sync
	//var test_wg sync.WaitGroup
	//test_wg.Add(1)
	//syncRegistrations(test_esl_client, []string{"internal"}, "192.168.99.100", 5061, 300, test_kv_backend, &test_wg, true)
	// TODO: check that the expected data is in etcd
}
