package main

import (
	"reflect"
	"testing"
)

func TestGenerateCurrentRegistrationsType(t *testing.T) {
	input := []string{
		"user1@domain",
		"user2@domain",
	}
	expected_result := Registrations{
		"user1@domain": KvBackendValue{
			Host: "10.20.30.40",
			Port: 5061,
		},
		"user2@domain": KvBackendValue{
			Host: "10.20.30.40",
			Port: 5061,
		},
	}
	result := generateCurrentRegistrationsType(&input, "10.20.30.40", 5061)
	if reflect.DeepEqual(*result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
}

func TestGenerateLastRegistrationsType(t *testing.T) {
	expected_result := Registrations{
		"user3@domain": KvBackendValue{
			Host: "10.20.30.50",
			Port: 5062,
		},
		"user4@domain": KvBackendValue{
			Host: "10.20.30.50",
			Port: 5062,
		},
	}
	// Test a valid one
	result1, err := generateLastRegistrationsType(&map[string]string{
		"user3@domain": "{\"host\":\"10.20.30.50\",\"port\":5062}",
		"user4@domain": "{\"host\":\"10.20.30.50\",\"port\":5062}",
	})
	if err != nil {
		t.Fatal("Expected no error, got", err)
	}
	if reflect.DeepEqual(*result1, expected_result) != true {
		t.Error("Expected", expected_result, "got", result1)
	}
	// And a failure
	_, err = generateLastRegistrationsType(&map[string]string{
		"user3@domain": "{\"host\":\"10.20.30.50\"\"port\":5062}",
		"user4@domain": "{\"host\":\"10.20.30.50\",\"port\":5062}",
	})
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func TestGenerateRegistrationListForThisInstance(t *testing.T) {
	expected_result := Registrations{
		"user5@domain": KvBackendValue{
			Host: "10.20.30.60",
			Port: 5063,
		},
		"user6@domain": KvBackendValue{
			Host: "10.20.30.60",
			Port: 5063,
		},
	}
	result := generateRegistrationListForThisInstance(&Registrations{
		"user3@domain": KvBackendValue{
			Host: "10.20.30.50",
			Port: 5062,
		},
		"user5@domain": KvBackendValue{
			Host: "10.20.30.60",
			Port: 5063,
		},
		"user4@domain": KvBackendValue{
			Host: "10.20.30.50",
			Port: 5062,
		},
		"user6@domain": KvBackendValue{
			Host: "10.20.30.60",
			Port: 5063,
		},
	}, "10.20.30.60", 5063)
	if reflect.DeepEqual(*result, expected_result) != true {
		t.Error("Expected", expected_result, "got", result)
	}
}

func TestReconcileRegistrations(t *testing.T) {
	// TODO: implement
}
