package main

import (
	"fmt"
	"github.com/0x19/goesl"
	"github.com/urfave/cli"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

func watchForRegistrationEvents(esl_client *goesl.Client, advertise_ip string, advertise_port int, kv_backend KvBackend, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("watchForRegistrationEvents(): Starting.\n")
	err := subscribeToFreeswitchRegEvents(esl_client)
	if err != nil {
		// TODO: log to an error channel?
		log.Fatal(err)
	}
	log.Printf("watchForRegistrationEvents(): Started.\n")
	for {
		msg, err := esl_client.ReadMessage()
		if err != nil {
			// If it contains EOF, we really dont care...
			if !strings.Contains(err.Error(), "EOF") && err.Error() != "unexpected end of JSON input" {
				log.Printf("(Ignored) Error while reading FreeSWITCH message: %s", err)
				continue
			}
			log.Printf("Error while reading FreeSWITCH message: %s", err)
			break
		}
		log.Printf("watchForRegistrationEvents() : New Message from FreeSWITCH - %+v\n", msg)
		reg_event, reg_event_user, err := parseFreeswitchRegEvent(msg)
		if err != nil {
			// TODO: log to an error channel?
			log.Fatal(err)
		}
		log.Printf("watchForRegistrationEvents() : Event - %s, User - %s\n", reg_event, reg_event_user)
	}
	log.Printf("watchForRegistrationEvents(): Finished.\n")
}

func syncRegistrations(esl_client *goesl.Client, sofia_profiles []string, advertise_ip string, advertise_port int, sync_interval uint32, kv_backend KvBackend, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		log.Printf("syncRegistrations(): Starting.\n")

		fs_registrations, err := getFreeswitchRegistrations(esl_client, sofia_profiles)
		if err != nil {
			// TODO: return an error channel or something?
			log.Fatal(err)
		}
		fmt.Printf("FS Registrations: %+v\n", fs_registrations)

		last_active_registrations, err := kv_backend.Read("", true)
		if err != nil {
			if err.Error() == "KEY_NOT_FOUND" {
				log.Printf("No active registrations found within K/V backend. Clean slate.\n")
			} else {
				log.Fatalf("Error reading from K/V Backend: %s\n", err)
			}
		}
		log.Printf("last_active_registrations: %+v\n", last_active_registrations)

		// Sleep between syncs, this is run in a goroutine.
		log.Printf("syncRegistrations(): Finished, sleeping for %d seconds.\n", sync_interval)
		time.Sleep(time.Duration(sync_interval) * time.Second)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "fs-registrator"
	app.Version = "0.1.0"
	app.Usage = "FreeSWITCH Sofia-SIP Registry Bridge (Sync to Key/Value Store)"
	app.Action = func(c *cli.Context) error {
		arg_config := parseFlags(c)

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
			Usage: fmt.Sprintf("Key/Value Backend (one of: %s)", strings.Join(availableKvBackends(), ", ")),
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
