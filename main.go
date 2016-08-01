package main

import (
	"github.com/0x19/goesl"
	"github.com/urfave/cli"
	"log"
	"os"
	"strings"
	"sync"
)

type ArgConfig struct {
	// FreeSWITCH
	FreeswitchHost          string
	FreeswitchPort          uint
	FreeswitchEslPassword   string
	FreeswitchSofiaProfiles []string
	FreeswitchAdvertiseIp   string
	FreeswitchAdvertisePort uint
	// Key/Value Store
	KvBackend string
	KvHost    string
	KvPort    uint
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
		if uint(c.Int(v)) <= 0 {
			log.Printf("Error: --%s must not be 0 (or empty).\n\n", v)
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
	}
	result.FreeswitchHost = c.String("fshost")
	result.FreeswitchPort = uint(c.Int("fsport"))
	result.FreeswitchEslPassword = c.String("fspassword")
	result.KvHost = c.String("kvhost")
	result.KvPort = uint(c.Int("kvport"))
	result.KvPrefix = c.String("kvprefix")

	// Add more here as they are supported.
	available_backends := []string{"etcd"}
	if stringInSlice(c.String("kvbackend"), available_backends) != true {
		log.Printf("Error: --kvbackend must be one of: %s\n\n", strings.Join(available_backends, ", "))
		cli.ShowAppHelp(c)
		os.Exit(1)
	}

	if uint32(c.Int("syncinterval")) <= 0 {
		log.Printf("Error: --syncinterval must not be 0 (or empty).\n\n")
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	result.SyncInterval = uint32(c.Int("syncinterval"))

	result.FreeswitchSofiaProfiles = strings.Split(c.String("fsprofiles"), ",")

	return &result
}

func main() {
	app := cli.NewApp()
	app.Name = "fs-registrator"
	app.Version = "0.1.0"
	app.Usage = "FreeSWITCH Sofia-SIP Registry Bridge (Sync to Key/Value Store)"
	app.Action = func(c *cli.Context) error {
		arg_config := parseFlags(c)

		log.Printf("Opening FreeSWITCH ESL Connections (%s:%d)...", arg_config.FreeswitchHost, arg_config.FreeswitchPort)
		// TODO: reconnection attempts? or just exit?
		event_client, err := goesl.NewClient(arg_config.FreeswitchHost, arg_config.FreeswitchPort, arg_config.FreeswitchEslPassword, int(5))
		if err != nil {
			log.Fatal(err)
		}
		// TODO: reconnection attempts? or just exit?
		sync_client, err := goesl.NewClient(arg_config.FreeswitchHost, uint(arg_config.FreeswitchPort), arg_config.FreeswitchEslPassword, int(5))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("FreeSWITCH ESL Connections Established.")

		// Setup our KV backend client.
		var kv_backend KvBackend
		switch arg_config.KvBackend {
		case "etcd":
			var kv_backend KvBackendEtcd
			kv_backend.SetupEtcdClient(arg_config.KvHost, arg_config.KvPort)
		}

		var wg sync.WaitGroup

		go event_client.Handle()
		go sync_client.Handle()
		wg.Add(1)
		go watchForRegistrationEvents(&event_client, &kv_backend, &wg)
		wg.Add(1)
		go syncRegistrations(&sync_client, arg_config.FreeswitchSofiaProfiles, arg_config.SyncInterval, &kv_backend, &wg)

		wg.Wait()

		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "fshost",
			Value: "localhost",
			Usage: "FreeSWITCH ESL Hostname/IP",
		},
		cli.IntFlag{
			Name:  "fsport",
			Value: 8021,
			Usage: "FreeSWITCH ESL Port",
		},
		cli.StringFlag{
			Name:  "fspassword",
			Value: "ClueCon",
			Usage: "FreeSWITCH ESL Password",
		},
		cli.StringFlag{
			Name:  "fsprofiles",
			Value: "internal",
			Usage: "List of Sofia Profiles to watch (comma separated list)",
		},
		cli.StringFlag{
			Name:  "fsadvertiseip",
			Value: "",
			Usage: "SIP Destination IP to store in K/V Store for FreeSWITCH",
		},
		cli.StringFlag{
			Name:  "fsadvertiseport",
			Value: "",
			Usage: "SIP Destination Port to store in K/V Store for FreeSWITCH",
		},
		cli.StringFlag{
			Name:  "kvbackend",
			Value: "etcd",
			Usage: "Key/Value Backend (one of 'etcd')",
		},
		cli.StringFlag{
			Name:  "kvhost",
			Value: "etcd",
			Usage: "Key/Value Store Hostname/IP",
		},
		cli.IntFlag{
			Name:  "kvport",
			Value: 2379,
			Usage: "Key/Value Store Port",
		},
		cli.StringFlag{
			Name:  "kvprefix",
			Value: "fs_registrations",
			Usage: "Key Space Prefix in K/V Store to store Registrations",
		},
		cli.IntFlag{
			Name:  "syncinterval",
			Value: 3600,
			Usage: "Interval (in seconds) between full sync. A full sync is performed on initial startup also.",
		},
	}

	app.Run(os.Args)
}
