package main

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/urfave/cli.v1"
)

type ArgConfig struct {
	// FreeSWITCH
	FreeswitchHost          string
	FreeswitchPort          int
	FreeswitchEslPassword   string
	FreeswitchSofiaProfiles []string
	FreeswitchAdvertiseIp   string
	FreeswitchAdvertisePort int
	// Key/Value Store
	KvBackend string
	KvHost    string
	KvPort    int
	KvPrefix  string
	//
	SyncInterval uint32
}

func parseFlags(c *cli.Context) (*ArgConfig, error) {
	var result ArgConfig

	for _, v := range []string{"fshost", "fspassword", "fsprofiles", "fsadvertiseip", "kvhost", "kvprefix"} {
		if len(c.String(v)) == 0 {
			return new(ArgConfig), errors.New(fmt.Sprintf("Error: --%s must not be empty.", v))
		}
	}
	for _, v := range []string{"fsport", "fsadvertiseport", "kvport"} {
		if c.Int(v) <= 0 {
			return new(ArgConfig), errors.New(fmt.Sprintf("Error: --%s must not be 0 (or empty).", v))
		}
		if c.Int(v) > 65536 {
			return new(ArgConfig), errors.New(fmt.Sprintf("Error: --%s must be below 65536.", v))
		}
	}
	result.FreeswitchHost = c.String("fshost")
	result.FreeswitchPort = c.Int("fsport")
	result.FreeswitchEslPassword = c.String("fspassword")
	result.FreeswitchAdvertiseIp = c.String("fsadvertiseip")
	result.FreeswitchAdvertisePort = c.Int("fsadvertiseport")
	result.KvHost = c.String("kvhost")
	result.KvPort = c.Int("kvport")
	result.KvPrefix = c.String("kvprefix")

	available_backends := availableKvBackends()
	if stringInSlice(c.String("kvbackend"), available_backends) != true {
		return new(ArgConfig), errors.New(fmt.Sprintf("Error: --kvbackend must be one of: %s", strings.Join(available_backends, ", ")))
	}
	result.KvBackend = c.String("kvbackend")

	if uint32(c.Int("syncinterval")) <= 0 {
		return new(ArgConfig), errors.New("Error: --syncinterval must not be 0 (or empty).")
	}
	result.SyncInterval = uint32(c.Int("syncinterval"))

	result.FreeswitchSofiaProfiles = strings.Split(c.String("fsprofiles"), ",")

	return &result, nil
}
