package main

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/urfave/cli.v1"
)

func TestParseFlags(t *testing.T) {
	expected_result1 := new(ArgConfig)
	expected_result1.FreeswitchHost = "somehost"
	expected_result1.FreeswitchPort = 8022
	expected_result1.FreeswitchEslPassword = "somepass"
	expected_result1.FreeswitchSofiaProfiles = []string{"profile1", "profile2"}
	expected_result1.FreeswitchAdvertiseIp = "10.3.4.5"
	expected_result1.FreeswitchAdvertisePort = 5071
	expected_result1.KvBackend = "etcd"
	expected_result1.KvHost = "somekvhost"
	expected_result1.KvPort = 2380
	expected_result1.KvPrefix = "someprefix"
	expected_result1.SyncInterval = 330

	set1 := flag.NewFlagSet("test1", 0)
	set1.String("fshost", "somehost", "doc")
	set1.Int("fsport", 8022, "doc")
	set1.String("fspassword", "somepass", "doc")
	set1.String("fsprofiles", "profile1,profile2", "doc")
	set1.String("fsadvertiseip", "10.3.4.5", "doc")
	set1.Int("fsadvertiseport", 5071, "doc")
	set1.String("kvbackend", "etcd", "doc")
	set1.String("kvhost", "somekvhost", "doc")
	set1.Int("kvport", 2380, "doc")
	set1.String("kvprefix", "someprefix", "doc")
	set1.Int("syncinterval", 330, "doc")
	context1 := cli.NewContext(nil, set1, nil)

	result1, err := parseFlags(context1)
	if err != nil {
		t.Error("Expected nil error, got", err)
	}
	if reflect.DeepEqual(result1, expected_result1) != true {
		t.Error("Expected", result1, "got", expected_result1)
	}

	// Cover some of the error scenarios.
	set2 := flag.NewFlagSet("test2", 0)
	set2.String("fshost", "", "doc")
	context2 := cli.NewContext(nil, set2, nil)
	_, err = parseFlags(context2)
	if err == nil {
		t.Error("Expected error, got nil error")
	}
	expected_err2 := "Error: --fshost must not be empty."
	if err.Error() != expected_err2 {
		t.Error("Expected error of", expected_err2, "got", err.Error())
	}
	//
	set3 := flag.NewFlagSet("test3", 0)
	set3.String("fshost", "10.20.30.40", "doc")
	set3.String("fspassword", "somepass", "doc")
	set3.String("fsprofiles", "profile1,profile2", "doc")
	set3.String("fsadvertiseip", "10.3.4.5", "doc")
	set3.String("kvhost", "somekvhost", "doc")
	set3.String("kvprefix", "someprefix", "doc")
	set3.Int("fsport", 0, "doc")
	context3 := cli.NewContext(nil, set3, nil)
	_, err = parseFlags(context3)
	if err == nil {
		t.Error("Expected error, got nil error")
	}
	expected_err3 := "Error: --fsport must not be 0 (or empty)."
	if err.Error() != expected_err3 {
		t.Error("Expected error of", expected_err3, "got", err.Error())
	}
	//
	set4 := flag.NewFlagSet("test4", 0)
	set4.String("fshost", "10.20.30.40", "doc")
	set4.String("fspassword", "somepass", "doc")
	set4.String("fsprofiles", "profile1,profile2", "doc")
	set4.String("fsadvertiseip", "10.3.4.5", "doc")
	set4.String("kvhost", "somekvhost", "doc")
	set4.String("kvprefix", "someprefix", "doc")
	set4.Int("fsport", 67000, "doc")
	context4 := cli.NewContext(nil, set4, nil)
	_, err = parseFlags(context4)
	if err == nil {
		t.Error("Expected error, got nil error")
	}
	expected_err4 := "Error: --fsport must be below 65536."
	if err.Error() != expected_err4 {
		t.Error("Expected error of", expected_err4, "got", err.Error())
	}
	//
	set5 := flag.NewFlagSet("test1", 0)
	set5.String("fshost", "somehost", "doc")
	set5.Int("fsport", 8022, "doc")
	set5.String("fspassword", "somepass", "doc")
	set5.String("fsprofiles", "profile1,profile2", "doc")
	set5.String("fsadvertiseip", "10.3.4.5", "doc")
	set5.Int("fsadvertiseport", 5071, "doc")
	set5.String("kvhost", "somekvhost", "doc")
	set5.Int("kvport", 2380, "doc")
	set5.String("kvprefix", "someprefix", "doc")
	set5.String("kvbackend", "randombackend", "doc")
	context5 := cli.NewContext(nil, set5, nil)
	_, err = parseFlags(context5)
	if err == nil {
		t.Error("Expected error, got nil error")
	}
	expected_err5 := fmt.Sprintf("Error: --kvbackend must be one of: %s", strings.Join(availableKvBackends(), ", "))
	if err.Error() != expected_err5 {
		t.Error("Expected error of", expected_err5, "got", err.Error())
	}
	//
	set6 := flag.NewFlagSet("test1", 0)
	set6.String("fshost", "somehost", "doc")
	set6.Int("fsport", 8022, "doc")
	set6.String("fspassword", "somepass", "doc")
	set6.String("fsprofiles", "profile1,profile2", "doc")
	set6.String("fsadvertiseip", "10.3.4.5", "doc")
	set6.Int("fsadvertiseport", 5071, "doc")
	set6.String("kvhost", "somekvhost", "doc")
	set6.Int("kvport", 2380, "doc")
	set6.String("kvprefix", "someprefix", "doc")
	set6.String("kvbackend", "etcd", "doc")
	set6.Int("syncinterval", 0, "doc")
	context6 := cli.NewContext(nil, set6, nil)
	_, err = parseFlags(context6)
	if err == nil {
		t.Error("Expected error, got nil error")
	}
	expected_err6 := "Error: --syncinterval must not be 0 (or empty)."
	if err.Error() != expected_err6 {
		t.Error("Expected error of", expected_err6, "got", err.Error())
	}
}
