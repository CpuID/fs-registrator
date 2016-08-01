package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/0x19/goesl"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
	"log"
	"strings"
	"sync"
	"time"
)

func watchForRegistrationEvents(esl_client *goesl.Client, kv_backend *KvBackend, wg *sync.WaitGroup) {
	defer wg.Done()
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
			break
		}
		log.Printf("New Message from FreeSWITCH: %+v\n", msg)
	}
	log.Printf("watchForRegistrationEvents(): Finished.\n")
}

type FsRegProfile struct {
	Profile       xml.Name                   `xml:"profile"`
	Registrations []FsRegProfileRegistration `xml:"registrations>registration"`
}

type FsRegProfileRegistration struct {
	CallId       string  `xml:"call-id"`
	User         string  `xml:"user"`
	Contact      string  `xml:"contact"`
	Agent        string  `xml:"agent"`
	Status       string  `xml:"status"`
	PingStatus   string  `xml:"ping-status"`
	PingTime     float64 `xml:"ping-time"`
	Host         string  `xml:"host"`
	NetworkIp    string  `xml:"network-ip"`
	NetworkPort  string  `xml:"network-port"`
	SipAuthUser  string  `xml:"sip-auth-user"`
	SipAuthRealm string  `xml:"sip-auth-realm"`
	MwiAccount   string  `xml:"mwi-account"`
}

// TODO: some kind of return dataset.
func getFreeswitchRegistrations(esl_client *goesl.Client, sofia_profiles []string) {
	for _, sofia_profile := range sofia_profiles {
		log.Printf("getFreeswitchRegistrations(): Fetching Registrations for Sofia Profile '%s'.\n", sofia_profile)
		esl_client.Send(fmt.Sprintf("api sofia xmlstatus profile %s reg", sofia_profile))
		msg, err := esl_client.ReadMessage()
		if err != nil {
			// TODO: decide on the right course of action here...
			// If it contains EOF, we really dont care...
			if !strings.Contains(err.Error(), "EOF") && err.Error() != "unexpected end of JSON input" {
				log.Printf("Error while reading Freeswitch message: %s", err.Error())
				continue
			}
			// TODO: return with an error channel instead?
			log.Fatal(err)
		}
		// TODOLATER: do we want to check the msg.Headers at all?
		var parsed_msg FsRegProfile
		// The XML is ISO-8859-1 as received from FreeSWITCH, convert to UTF-8.
		decoder := xml.NewDecoder(bytes.NewBuffer(msg.Body))
		decoder.CharsetReader = charset.NewReader
		err = decoder.Decode(&parsed_msg)
		if err != nil {
			// TODO: return with an error channel instead?
			log.Fatal(err)
		}
		log.Printf("Sofia Profile '%s' Registrations: %+v\n", sofia_profile, parsed_msg)
		// TODO: parse out msg, and reconcile against K/V store data.
	}
}

func syncRegistrations(esl_client *goesl.Client, sofia_profiles []string, sync_interval uint32, kv_backend *KvBackend, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		log.Printf("syncRegistrations(): Starting.\n")

		getFreeswitchRegistrations(esl_client, sofia_profiles)

		//last_active_registrations := kv_backend.Read("")

		// Sleep between syncs, this is run in a goroutine.
		log.Printf("syncRegistrations(): Finished, sleeping for %d seconds.\n", sync_interval)
		time.Sleep(time.Duration(sync_interval) * time.Second)
	}
}
