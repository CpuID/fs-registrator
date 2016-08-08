package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	//"reflect"
	"testing"

	"github.com/0x19/goesl"
)

func checkSipPortIsAvailable(t *testing.T) {
	if _, ok := dockerContainerPorts["freeswitch_1-5060/udp"]; ok == false {
		t.Fatal("Docker Container port for FreeSWITCH SIP not found in dockerContainerPorts, did the container start?")
	}
	log.Printf("checkSipPortIsAvailable() : Docker Container FreeSWITCH SIP Port - %d\n", uint(dockerContainerPorts["freeswitch_1-5060/udp"]))
}

func simulateSipRegister(host string, port uint, user string, password string, contact_port uint, t *testing.T) error {
	// sipsak -U -d -n -x 120 -C "sip:username@127.0.0.1:49201" -s "sip:username@192.168.99.100" -vvv -a somepassword
	cmd := exec.Command("sipsak", "-U", "-d", "-n", "-x", "120", "-C", fmt.Sprintf("sip:%s@127.0.0.1:%d", user, contact_port), "-s", fmt.Sprintf("sip:%s@%s:%d", user, host, port), "-vvv", "-a", password)
	//log.Printf("simulateSipRegister() : Command - %+v\n", cmd)
	_, err := cmd.CombinedOutput()
	// If SIP message fails to get a 200 OK back, a non-zero exit code will be returned.
	if err != nil {
		log.Printf("simulateSipRegister() : Command (that errored) - %+v\n", cmd)
		t.Fatal(err)
	}
	//log.Printf("SIP Register Output: %s\n", out)
	return nil
}

func simulateSipDeregister(host string, port uint, user string, password string, contact_port uint, t *testing.T) error {
	// sipsak -U -d -n -x 0 -C "<sip:username@127.0.0.1:49201>;expires=0" -s "sip:username@192.168.99.100" -vvv -a somepassword
	cmd := exec.Command("sipsak", "-U", "-d", "-n", "-x", "0", "-C", fmt.Sprintf("<sip:%s@127.0.0.1:%d>;expires=0", user, contact_port), "-s", fmt.Sprintf("sip:%s@%s:%d", user, host, port), "-vvv", "-a", password)
	//log.Printf("simulateSipDeregister() : Command - %+v\n", cmd)
	_, err := cmd.CombinedOutput()
	// If SIP message fails to get a 200 OK back, a non-zero exit code will be returned.
	if err != nil {
		log.Printf("simulateSipDeregister() : Command (that errored) - %+v\n", cmd)
		t.Fatal(err)
	}
	//log.Printf("SIP Deregister Output: %s\n", out)
	return nil
}

func getTestEslClient(t *testing.T) *goesl.Client {
	if _, ok := dockerContainerPorts["freeswitch_1-8021/tcp"]; ok == false {
		t.Fatal("Docker Container port for FreeSWITCH ESL not found in dockerContainerPorts, did the container start?")
	}
	log.Printf("getTestEslClient() : Docker Container FreeSWITCH ESL Port - %d\n", uint(dockerContainerPorts["freeswitch_1-8021/tcp"]))
	test_client, err := goesl.NewClient(dockerHost, uint(dockerContainerPorts["freeswitch_1-8021/tcp"]), "ClueCon", int(5))
	if err != nil {
		t.Fatal(err)
	}
	go test_client.Handle()
	return &test_client
}

/*
// TODO: we may have to rely on the tests in goroutine_test.go for this one,
// as its blocking and would need to be run in a goroutine otherwise. could possibly do it with channels standalone...
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
	// The single registration that occurred above will look like 1000@x.x.x.x, with the IP being based on the Docker network in use.
	// This is commonly a 172.17.0.x address, let's just regex search for an IP.
	expected_re := regexp.MustCompile("^1000@[0-9]+.[0-9]+.[0-9]+.[0-9]+$")
	//
	test_client := getTestEslClient(t)
	checkSipPortIsAvailable(t)
	simulateSipRegister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), "1000", "1234", uint(49201), t)
	result, err := getFreeswitchRegistrations(test_client, []string{"internal"})
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
	//log.Printf("Test FS Registrations: %+v\n", result)
	if len(*result) != 1 {
		t.Error("Expected 1 registration, got", len(*result))
	}
	result_no_pointer := *result
	if expected_re.MatchString(result_no_pointer[0]) != true {
		t.Error("Expected single registration to match regex", expected_re.String(), "- no match")
	}
	// Cleanup so other tests can make registrations if required.
	simulateSipDeregister(dockerHost, uint(dockerContainerPorts["freeswitch_1-5060/udp"]), "1000", "1234", uint(49201), t)
}
