package main

import (
	"flag"
	"reflect"
	"testing"

	"gopkg.in/urfave/cli.v1"
)

func TestParseFlags(t *testing.T) {
	expected_result := new(ArgConfig)
	expected_result.FreeswitchHost = "somehost"
	expected_result.FreeswitchPort = 8022
	expected_result.FreeswitchEslPassword = "somepass"
	expected_result.FreeswitchSofiaProfiles = []string{"profile1", "profile2"}
	expected_result.FreeswitchAdvertiseIp = "10.3.4.5"
	expected_result.FreeswitchAdvertisePort = 5071
	expected_result.KvBackend = "etcd"
	expected_result.KvHost = "somekvhost"
	expected_result.KvPort = 2380
	expected_result.KvPrefix = "someprefix"
	expected_result.SyncInterval = 330

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

	result := parseFlags(context1)

	if reflect.DeepEqual(result, expected_result) != true {
		t.Error("Expected", result, "got", expected_result)
	}
}
