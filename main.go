package main

import (
	"fmt"
	"github.com/0x19/goesl"
	"github.com/kr/pretty"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	app := cli.NewApp()
	app.Name = "fs-registrator"
	app.Version = "0.1.0"
	app.Usage = "FreeSWITCH Sofia-SIP Registry Bridge (Sync to Key/Value Store)"
	app.Action = func(c *cli.Context) error {
		arg_config := parseFlags(c)
		log.Printf("Config: %# v\n", pretty.Formatter(arg_config))

		// Setup our KV backend client.
		log.Printf("Setting up K/V (%s) Backend...", arg_config.KvBackend)
		kv_backend, err := CreateKvBackend(map[string]string{
			"backend": arg_config.KvBackend,
			"host":    arg_config.KvHost,
			"port":    strconv.Itoa(int(arg_config.KvPort)),
			"prefix":  arg_config.KvPrefix,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("K/V Backend Ready.\n")

		log.Printf("Opening FreeSWITCH ESL Connections (%s:%d)...", arg_config.FreeswitchHost, arg_config.FreeswitchPort)
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

		var wg sync.WaitGroup

		go event_client.Handle()
		go sync_client.Handle()
		wg.Add(1)
		go watchForRegistrationEvents(&event_client, arg_config.FreeswitchAdvertiseIp, arg_config.FreeswitchAdvertisePort, kv_backend, &wg)
		wg.Add(1)
		go syncRegistrations(&sync_client, arg_config.FreeswitchSofiaProfiles, arg_config.FreeswitchAdvertiseIp, arg_config.FreeswitchAdvertisePort, arg_config.SyncInterval, kv_backend, &wg)

		wg.Wait()

		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "fshost",
			Value:  "localhost",
			Usage:  "FreeSWITCH ESL Hostname/IP",
			EnvVar: "FS_HOST",
		},
		cli.IntFlag{
			Name:   "fsport",
			Value:  8021,
			Usage:  "FreeSWITCH ESL Port",
			EnvVar: "FS_PORT",
		},
		cli.StringFlag{
			Name:   "fspassword",
			Value:  "ClueCon",
			Usage:  "FreeSWITCH ESL Password",
			EnvVar: "FS_PASSWORD",
		},
		cli.StringFlag{
			Name:   "fsprofiles",
			Value:  "internal",
			Usage:  "List of Sofia Profiles to watch (comma separated list)",
			EnvVar: "FS_PROFILES",
		},
		cli.StringFlag{
			Name:   "fsadvertiseip",
			Value:  "",
			Usage:  "SIP Destination IP to store in K/V Store for FreeSWITCH",
			EnvVar: "FS_ADVERTISE_IP",
		},
		cli.StringFlag{
			Name:   "fsadvertiseport",
			Value:  "",
			Usage:  "SIP Destination Port to store in K/V Store for FreeSWITCH",
			EnvVar: "FS_ADVERTISE_PORT",
		},
		cli.StringFlag{
			Name:   "kvbackend",
			Value:  "etcd",
			Usage:  fmt.Sprintf("Key/Value Backend (one of: %s)", strings.Join(availableKvBackends(), ", ")),
			EnvVar: "KV_BACKEND",
		},
		cli.StringFlag{
			Name:   "kvhost",
			Value:  "etcd",
			Usage:  "Key/Value Store Hostname/IP",
			EnvVar: "KV_HOST",
		},
		cli.IntFlag{
			Name:   "kvport",
			Value:  2379,
			Usage:  "Key/Value Store Port",
			EnvVar: "KV_PORT",
		},
		cli.StringFlag{
			Name:   "kvprefix",
			Value:  "fs_registrations",
			Usage:  "Key Space Prefix in K/V Store to store Registrations",
			EnvVar: "KV_PREFIX",
		},
		cli.IntFlag{
			Name:   "syncinterval",
			Value:  3600,
			Usage:  "Interval (in seconds) between full sync. A full sync is performed on initial startup also.",
			EnvVar: "SYNC_INTERVAL",
		},
	}

	app.Run(os.Args)
}
