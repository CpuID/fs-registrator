package main

import (
	"fmt"
	"github.com/0x19/goesl"
	"log"
	"strings"
	"time"
)

// TODO: add a done channel so we can watch for it in main()
func watchForRegistrationEvents(esl_client *goesl.Client) {
	log.Printf("watchForRegistrationEvents(): Starting.\n")
	esl_client.Send("events json CUSTOM sofia::register sofia::expire")
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
			// TODO: end with the done channel
			break
		}
		log.Printf("New Message from FreeSWITCH: %+v\n", msg)
	}
	log.Printf("watchForRegistrationEvents(): Finished.\n")
}

// TODO: add a done channel so we can watch for it in main()
func syncRegistrations(esl_client *goesl.Client, sofia_profiles []string, sync_interval uint32) {
	for {
		log.Printf("syncRegistrations(): Starting.\n")

		for _, sofia_profile := range sofia_profiles {
			log.Printf("syncRegistrations(): Syncing Sofia Profile '%s'.\n", sofia_profile)
			esl_client.Send(fmt.Sprintf("api sofia xmlstatus profile %s reg", sofia_profile))
			msg, err := esl_client.ReadMessage()
			if err != nil {
				// TODO: decide on the right course of action here...
				// If it contains EOF, we really dont care...
				if !strings.Contains(err.Error(), "EOF") && err.Error() != "unexpected end of JSON input" {
					log.Printf("Error while reading Freeswitch message: %s", err.Error())
					continue
				}
				// TODO: return with a done channel instead?
				log.Fatal(err)
			}
			log.Printf("Sofia Profile '%s' Registrations: %+v\n", sofia_profile, msg)
			// TODO: parse out msg, and reconcile against K/V store data.
		}

		// Sleep between syncs, this is run in a goroutine.
		log.Printf("syncRegistrations(): Finished, sleeping for %d seconds.\n", sync_interval)
		time.Sleep(time.Duration(sync_interval) * time.Second)
	}
}
