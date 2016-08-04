package main

import (
	"reflect"
	"testing"
)

func TestAvailableKvBackends(t *testing.T) {
	expected_result := []string{
		"etcd",
		// Add new backends here as they become available.
	}
	result := availableKvBackends()
	if reflect.DeepEqual(result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
}

func TestCreateKvBackend(t *testing.T) {
	// Test a valid backend
	result, err := CreateKvBackend(map[string]string{
		"backend": "etcd",
		"host":    "10.2.3.4",
		"port":    "2379",
		"prefix":  "someprefix",
	})
	if err != nil {
		t.Fatal("Expected no error, got", err)
	}
	// No easy way to test equality on result.Kapi... its a private type upstream :(
	result_name := result.BackendName()
	if result_name != "etcd" {
		t.Error("Expected a return type of etcd, got", result_name)
	}
	result_prefix := result.GetPrefix()
	if result_prefix != "someprefix" {
		t.Error("Expected a .Prefix of someprefix, got", result_prefix)
	}
	// And a failure
	_, err = CreateKvBackend(map[string]string{
		"backend": "nonexistent",
	})
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func TestGetKvBackendValueType(t *testing.T) {
	expected_result := KvBackendValue{
		Host: "10.2.3.4",
		Port: 5061,
	}
	result := getKvBackendValueType("10.2.3.4", 5061)
	if reflect.DeepEqual(result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
}

func TestGetKvBackendValueJsonType(t *testing.T) {
	expected_result := KvBackendValue{
		Host: "10.3.4.5",
		Port: 5064,
	}
	// Test a valid entry
	result, err := getKvBackendValueJsonType("{\"host\":\"10.3.4.5\",\"port\":5064}")
	if err != nil {
		t.Fatal("Expected no error, got", err)
	}
	if reflect.DeepEqual(result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
	// And a failure
	_, err = getKvBackendValueJsonType("{\"host\"\"10.3.4.5\",\"port\":5064}")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func TestGetKvBackendValueJsonString(t *testing.T) {
	expected_result := "{\"host\":\"10.4.5.6\",\"port\":5065}"
	// Test a valid entry
	result, err := getKvBackendValueJsonString(KvBackendValue{
		Host: "10.4.5.6",
		Port: 5065,
	})
	if err != nil {
		t.Fatal("Expected no error, got", err)
	}
	if reflect.DeepEqual(result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
	// And a failure
	// TODOLATER: find something that will fail this, that still builds.
	/*
		_, err = getKvBackendValueJsonString(KvBackendValue{
			Host:      "10.4.5.6",
			Port:      5065,
		})
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
	*/
}
