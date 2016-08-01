package main

import (
	"github.com/0x19/goesl"
	"github.com/urfave/cli"
	"log"
	"os"
)

type ArgConfig struct {
	// Key/Value Store
	KvBackend string
	KvHost    string
	KvPort    uint8
	KvPrefix  string
	// FreeSWITCH
	FreeswitchHost          string
	FreeswitchPort          uint8
	FreeswitchEslPassword   string
	FreeswitchTimeout       uint8
	FreeswitchSofiaProfiles []string
	//
	SyncInterval uint32
}

func parseFlags(c *cli.Context) *ArgConfig {
	var result ArgConfig
	// TODO: parse stuff etc

	return &result
}

func main() {
	app := cli.NewApp()
	app.Name = "fs-registrator"
	app.Version = "0.1.0"
	app.Usage = "FreeSWITCH Sofia-SIP Registry Bridge (Sync to Key/Value Store)"
	app.Action = func(c *cli.Context) error {
		arg_config := parseFlags(c)

		log.Printf("Making FreeSWITCH ESL Connections...")
		// TODO: reconnection attempts? or just exit?
		event_client, err := goesl.NewClient(arg_config.FreeswitchHost, uint(arg_config.FreeswitchPort), arg_config.FreeswitchEslPassword, int(5))
		if err != nil {
			log.Fatal(err)
		}
		// TODO: reconnection attempts? or just exit?
		sync_client, err := goesl.NewClient(arg_config.FreeswitchHost, uint(arg_config.FreeswitchPort), arg_config.FreeswitchEslPassword, int(5))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("FreeSWITCH ESL Connections Established.")

		go event_client.Handle()
		go sync_client.Handle()
		go watchForRegistrationEvents(&event_client)
		go syncRegistrations(&sync_client, arg_config.FreeswitchSofiaProfiles, arg_config.SyncInterval)
		// TODO: handle some done channels for the above

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
