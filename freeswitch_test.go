package main

import (
	"log"
	"os/exec"
	//"reflect"
	"testing"

	"github.com/0x19/goesl"
)

func simulateSipRegister(user string, password string) error {
	// TODO: implement
	// sipsak -U -d -n -x 120 -C "sip:username@127.0.0.1:49201" -s "sip:username@192.168.99.100" -vvv -a nathans
	// out, err := exec.Command("sipsak", "-U", "-D", "", "").Output()
	return nil
}

func simulateSipDeregister(user string, password string) error {
	// TODO: implement
	// sipsak -U -d -n -x 0 -C "<sip:username@127.0.0.1:49201>;expires=0" -s "sip:username@192.168.99.100" -vvv -a nathans
	// out, err := exec.Command("sipsak", "-U", "-D", "", "").Output()
	return nil
}

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

/*
func TestSubscribeToFreeswitchRegEvents(t *testing.T) {
	test_client := getTestEslClient(t)
	err := subscribeToFreeswitchRegEvents(test_client)
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
}
*/

func TestParseFreeswitchRegEvent(t *testing.T) {
	expected_result1 := "register"
	expected_result2 := "someuser@sip.somedomain.com"
	input := goesl.Message{
		Headers: map[string]string{
			"call-id":                   "AbtneHy2nQkhY-S.ypzYrl25I9zEIPGN",
			"contact":                   "\"Firstname Lastname\" <sip:someuser@192.168.99.1:58843;ob>",
			"Core-UUID":                 "dbedeff1-2070-4bce-a320-9669ba067f02",
			"expires":                   "300",
			"Event-Calling-File":        "sofia_reg.c",
			"Event-Calling-Function":    "sofia_reg_handle_register_token",
			"Event-Calling-Line-Number": "2002",
			"Event-Date-GMT":            "Fri, 05 Aug 2016 03:17:51",
			"Event-Date-Local":          "2016-08-05 03:17:51",
			"Event-Date-Timestamp":      "1470367071126720",
			"Event-Name":                "CUSTOM",
			"Event-Sequence":            "478",
			"Event-Subclass":            "sofia::register",
			"FreeSWITCH-Hostname":       "6ff0e1f477a1",
			"FreeSWITCH-IPv4":           "172.17.0.14",
			"FreeSWITCH-IPv6":           "::1",
			"FreeSWITCH-Switchname":     "6ff0e1f477a1",
			"from-host":                 "sip.somedomain.com",
			"from-user":                 "someuser",
			"network-ip":                "192.168.99.1",
			"network-port":              "58843",
			"presence-hosts":            "n/a",
			"profile-name":              "someprofile",
			"realm":                     "192.168.99.100",
			"rpid":                      "unknown",
			"status":                    "Registered(UDP)",
			"to-host":                   "192.168.99.100",
			"to-user":                   "someuser",
			"username":                  "someuser",
			"user-agent":                "Telephone 1.1.7",
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
	//expected_result := []string{
	// TODO: fill in
	//}
	test_client := getTestEslClient(t)
	_, err := getFreeswitchRegistrations(test_client, []string{
	// TODO: fill in
	})
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
	/*
		TODO: only enable once we have data in expected_result and result, otherwise its permafail due to pointers.
		if reflect.DeepEqual(result, expected_result) != true {
			t.Error("Expected", expected_result, "got", &result)
		}
	*/
}
