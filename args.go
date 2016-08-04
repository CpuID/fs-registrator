package main

import (
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"strings"
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

func parseFlags(c *cli.Context) *ArgConfig {
	var result ArgConfig

	for _, v := range []string{"fshost", "fspassword", "fsprofiles", "fsadvertiseip", "kvhost", "kvprefix"} {
		if len(c.String(v)) == 0 {
			log.Printf("Error: --%s must not be empty.\n\n", v)
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
	}
	for _, v := range []string{"fsport", "fsadvertiseport", "kvport"} {
		if c.Int(v) <= 0 {
			log.Printf("Error: --%s must not be 0 (or empty).\n\n", v)
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		if c.Int(v) > 65536 {
			log.Printf("Error: --%s must be below 65536.\n\n", v)
			cli.ShowAppHelp(c)
			os.Exit(1)
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
		log.Printf("Error: --kvbackend must be one of: %s\n\n", strings.Join(available_backends, ", "))
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	result.KvBackend = c.String("kvbackend")

	if uint32(c.Int("syncinterval")) <= 0 {
		log.Printf("Error: --syncinterval must not be 0 (or empty).\n\n")
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	result.SyncInterval = uint32(c.Int("syncinterval"))

	result.FreeswitchSofiaProfiles = strings.Split(c.String("fsprofiles"), ",")

	return &result
}
