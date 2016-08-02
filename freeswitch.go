package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/0x19/goesl"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
	"log"
	"strings"
)

// event_type string, user string, err error
func getFreeswitchRegEvent(event map[string]string) (string, string, error) {
	// These events don't have the full <user> like we get showing registrations, build it from username and from-host.
	for _, v := range []string{"Event-Subclass", "username", "from-host"} {
		if _, ok := event[v]; ok == false {
			return "", "", errors.New(fmt.Sprintf("getFreeswitchRegEvent() : '%s' field does not exist in FreeSWITCH Event, must be present.", v))
		}
		if len(event[v]) == 0 {
			return "", "", errors.New(fmt.Sprintf("getFreeswitchRegEvent() : '%s' field cannot be empty in FreeSWITCH Event.", v))
		}
	}
	valid_event_subclasses := []string{"sofia::register", "sofia::expire", "sofia::unregister"}
	if stringInSlice(event["Event-Subclass"], valid_event_subclasses) == false {
		return "", "", errors.New(fmt.Sprintf("getFreeswitchRegEvent() : 'Event-Subclass' field must be one of: %s", strings.Join(valid_event_subclasses, ", ")))
	}
	return strings.Replace(event["Event-Subclass"], "sofia::", "", 1), fmt.Sprintf("%s@%s", event["username"], event["from-host"]), nil
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
func getFreeswitchRegistrations(esl_client *goesl.Client, sofia_profiles []string) ([]string, error) {
	var results []string
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
			return []string{}, err
		}
		// TODOLATER: do we want to check the msg.Headers at all?
		var parsed_msg FsRegProfile
		// The XML is ISO-8859-1 as received from FreeSWITCH, convert to UTF-8.
		decoder := xml.NewDecoder(bytes.NewBuffer(msg.Body))
		decoder.CharsetReader = charset.NewReader
		err = decoder.Decode(&parsed_msg)
		if err != nil {
			return []string{}, err
		}
		//log.Printf("Sofia Profile '%s' Registrations: %+v\n", sofia_profile, parsed_msg)
		for _, v := range parsed_msg.Registrations {
			if len(v.User) > 0 && stringInSlice(v.User, results) == false {
				results = append(results, v.User)
			}
		}
	}
	return results, nil
}
