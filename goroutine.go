package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/0x19/goesl"
)

// Both of the below functions are run within goroutines (in parallel) from main()

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
		if reg_event == "register" {
			kv_backend_value_string, err := getKvBackendValueJsonString(KvBackendValue{
				Host: advertise_ip,
				Port: advertise_port,
			})
			if err != nil {
				// TODO: log to an error channel?
				log.Fatal(err)
			}
			// TODO: move the TTL out to somewhere more reusable
			err = kv_backend.Write(reg_event_user, kv_backend_value_string, 300)
			if err != nil {
				// TODO: log to an error channel?
				log.Fatal(err)
			}
		} else if reg_event == "unregister" || reg_event == "expire" {
			err = kv_backend.Delete(reg_event_user)
			if err != nil {
				// TODO: log to an error channel?
				log.Fatal(err)
			}
		}
	}
	log.Printf("watchForRegistrationEvents(): Finished.\n")
}

func syncRegistrations(esl_client *goesl.Client, sofia_profiles []string, advertise_ip string, advertise_port int, sync_interval uint32, kv_backend KvBackend, wg *sync.WaitGroup, once bool) {
	defer wg.Done()
	for {
		log.Printf("syncRegistrations(): Starting.\n")

		raw_current_active_registrations, err := getFreeswitchRegistrations(esl_client, sofia_profiles)
		if err != nil {
			// TODO: return an error channel or something?
			log.Fatal(err)
		}
		log.Printf("FS Registrations: %+v\n", raw_current_active_registrations)

		raw_last_active_registrations, err := kv_backend.Read("", true)
		if err != nil {
			if err.Error() == "KEY_NOT_FOUND" {
				log.Printf("No active registrations found within K/V backend. Clean slate.\n")
			} else {
				log.Fatalf("Error reading from K/V Backend: %s\n", err)
			}
		}
		log.Printf("last_active_registrations: %+v\n", raw_last_active_registrations)

		last_active_registrations_typed, err := generateLastRegistrationsType(raw_last_active_registrations)
		if err != nil {
			log.Fatal(err)
		}
		last_active_registrations := generateRegistrationListForThisInstance(last_active_registrations_typed, advertise_ip, advertise_port)
		current_active_registrations := generateCurrentRegistrationsType(raw_current_active_registrations, advertise_ip, advertise_port)

		add_registrations, remove_registrations, err := reconcileRegistrations(advertise_ip, advertise_port, last_active_registrations, current_active_registrations)
		if err != nil {
			// TODO: return an error channel or something?
			log.Fatal(err)
		}

		for k_add, v_add := range *add_registrations {
			v_add_json_string, err := getKvBackendValueJsonString(v_add)
			if err != nil {
				// TODO: return an error channel or something?
				log.Fatal(err)
			}
			// TODO: move the TTL out to somewhere more reusable
			err = kv_backend.Write(k_add, v_add_json_string, 300)
		}
		for _, v_remove := range *remove_registrations {
			err = kv_backend.Delete(v_remove)
		}

		// Used for test suite, to only do a once-off sync.
		if once == true {
			log.Printf("syncRegistrations(): Once off mode enabled, finished.\n")
			return
		}

		// Sleep between syncs, this is run in a goroutine.
		log.Printf("syncRegistrations(): Finished, sleeping for %d seconds.\n", sync_interval)
		time.Sleep(time.Duration(sync_interval) * time.Second)
	}
}
