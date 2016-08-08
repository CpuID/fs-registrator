package main

import (
	"sort"
)

// Key = username@domain
// We use the <user> value from "sofia xmlstatus profile internal reg" to populate.
type Registrations map[string]KvBackendValue

// The format we receive from FreeSWITCH.
func generateCurrentRegistrationsType(users *[]string, advertise_ip string, advertise_port int) *Registrations {
	result := make(Registrations)
	for _, v := range *users {
		// TODO: duplicate user handling?
		result[v] = KvBackendValue{
			Host: advertise_ip,
			Port: advertise_port,
		}
	}
	return &result
}

// The format we receive from a K/V backend.
func generateLastRegistrationsType(input *map[string]string) (*Registrations, error) {
	result := make(Registrations)
	for k, v := range *input {
		parse_v, err := getKvBackendValueJsonType(v)
		if err != nil {
			return new(Registrations), err
		}
		result[k] = parse_v
	}
	return &result, nil
}

// Parses out multiple K/V backend result sets into just the user@domain list,
// and filter on this advertise IP and port only (this instance).
func generateRegistrationListForThisInstance(input *Registrations, advertise_ip string, advertise_port int) *Registrations {
	result := make(Registrations)
	for k, v := range *input {
		if v.Host != advertise_ip || v.Port != advertise_port {
			continue
		}
		result[k] = v
	}
	return &result
}

// add_registrations []string, remove_registrations []string
func reconcileRegistrations(last_active_registrations *Registrations, current_active_registrations *Registrations) (*[]string, *[]string, error) {
	var add_registrations []string
	var remove_registrations []string

	// First in one direction.
	for k1, _ := range *last_active_registrations {
		exists_in_current := false
		for k2, _ := range *current_active_registrations {
			if k1 == k2 {
				exists_in_current = true
				break
			}
		}
		if exists_in_current == false {
			remove_registrations = append(remove_registrations, k1)
		}
	}

	// And in reverse.
	for k3, _ := range *current_active_registrations {
		exists_in_last := false
		for k4, _ := range *last_active_registrations {
			if k3 == k4 {
				exists_in_last = true
				break
			}
		}
		if exists_in_last == false {
			add_registrations = append(add_registrations, k3)
		}
	}

	// Reconcile our adds and removes, if we find any that are "added" and also "removed", they cancel eachother out and can be removed from both.
	// This is normally a sign they already exist.
	var add_registrations_results []string
	var remove_registrations_results []string
	for _, v5 := range add_registrations {
		if stringInSlice(v5, remove_registrations) == false {
			add_registrations_results = append(add_registrations_results, v5)
		}
	}
	for _, v6 := range remove_registrations {
		if stringInSlice(v6, add_registrations) == false {
			remove_registrations_results = append(remove_registrations_results, v6)
		}
	}

	// Sort the results
	sort.Strings(add_registrations_results)
	sort.Strings(remove_registrations_results)

	return &add_registrations_results, &remove_registrations_results, nil
}
