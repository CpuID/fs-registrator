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
	// Scenario 1, all adds, no removes.
	last_input1 := Registrations{}
	current_input1 := Registrations{
		"1002@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.60",
			Port: 5070,
		},
		"1003@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.60",
			Port: 5070,
		},
	}
	expected_add1 := []string{
		"1002@sip.testserver.tld",
		"1003@sip.testserver.tld",
	}
	expected_remove1 := []string{}
	result_add1, result_remove1, err := reconcileRegistrations(&last_input1, &current_input1)
	if err != nil {
		t.Error("Scenario 1: Expected nil error, got error", err)
	}
	if len(*result_add1) > 0 || len(expected_add1) > 0 {
		if reflect.DeepEqual(*result_add1, expected_add1) != true {
			t.Error("Scenario 1: Expected", expected_add1, "got", result_add1)
		}
	}
	if len(*result_remove1) > 0 || len(expected_remove1) > 0 {
		if reflect.DeepEqual(*result_remove1, expected_remove1) != true {
			t.Error("Scenario 1: Expected", expected_remove1, "got", result_remove1)
		}
	}

	// Scenario 2, some adds and removes in one operation.
	last_input2 := Registrations{
		"1002@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1003@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		// Out of order test
		"1010@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1005@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1006@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1008@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1009@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
	}
	current_input2 := Registrations{
		"1002@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1004@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		// Out of order test
		"1010@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1005@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1006@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1007@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
		"1008@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.70",
			Port: 5071,
		},
	}
	expected_add2 := []string{
		"1004@sip.testserver.tld",
		"1007@sip.testserver.tld",
	}
	expected_remove2 := []string{
		"1003@sip.testserver.tld",
		"1009@sip.testserver.tld",
	}
	result_add2, result_remove2, err := reconcileRegistrations(&last_input2, &current_input2)
	if err != nil {
		t.Error("Scenario 2: Expected nil error, got error", err)
	}
	if len(*result_add2) > 0 || len(expected_add2) > 0 {
		if reflect.DeepEqual(*result_add2, expected_add2) != true {
			t.Error("Scenario 2: Expected", expected_add2, "got", result_add2)
		}
	}
	if len(*result_remove2) > 0 || len(expected_remove2) > 0 {
		if reflect.DeepEqual(*result_remove2, expected_remove2) != true {
			t.Error("Scenario 2: Expected", expected_remove2, "got", result_remove2)
		}
	}

	// Scenario 3, all removes.
	last_input3 := Registrations{
		"1011@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.80",
			Port: 5072,
		},
		"1012@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.80",
			Port: 5072,
		},
		"1013@sip.testserver.tld": KvBackendValue{
			Host: "10.20.30.80",
			Port: 5072,
		},
	}
	current_input3 := Registrations{}
	expected_add3 := []string{}
	expected_remove3 := []string{
		"1011@sip.testserver.tld",
		"1012@sip.testserver.tld",
		"1013@sip.testserver.tld",
	}
	result_add3, result_remove3, err := reconcileRegistrations(&last_input3, &current_input3)
	if err != nil {
		t.Error("Scenario 3: Expected nil error, got error", err)
	}
	if len(*result_add3) > 0 || len(expected_add3) > 0 {
		if reflect.DeepEqual(*result_add3, expected_add3) != true {
			t.Error("Scenario 3: Expected", expected_add3, "got", result_add3)
		}
	}
	if len(*result_remove3) > 0 || len(expected_remove3) > 0 {
		if reflect.DeepEqual(*result_remove3, expected_remove3) != true {
			t.Error("Scenario 3: Expected", expected_remove3, "got", result_remove3)
		}
	}
}
