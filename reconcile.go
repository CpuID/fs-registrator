package main

import ()

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
func generateLastRegistrationsType(input map[string]string) (*Registrations, error) {
	var result Registrations
	for k, v := range input {
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
func generateRegistrationListForThisInstance(input Registrations, advertise_ip string, advertise_port int) *Registrations {
	var result Registrations
	for k, v := range input {
		if v.Host == advertise_ip && v.Port == advertise_port {
			continue
		}
		result[k] = v
	}
	return &result
}

// TODO: do we need to be reviewing advertise_ip and advertise_port here? or do we assume that this is pre-processed before we get to reconciliation?
// add_registrations Registrations, remove_registrations []string
func reconcileRegistrations(advertise_ip string, advertise_port int, last_active_registrations *Registrations, current_active_registrations *Registrations) (*Registrations, *[]string, error) {
	add_registrations := make(Registrations)
	var remove_registrations []string

	// TODO: we should only remove if the last registration was for this advertise IP.

	// First in one direction.
	for k1, _ := range *last_active_registrations {
		name_exists_in_current := false
		for k2, _ := range *current_active_registrations {
			if k1 == k2 {
				name_exists_in_current = true
				break
			}
		}
		if name_exists_in_current == false {
			remove_registrations = append(remove_registrations, k1)
		}
	}

	// And in reverse.
	for k3, v3 := range *current_active_registrations {
		exists_in_last := false
		for k4, _ := range *last_active_registrations {
			if k3 == k4 {
				exists_in_last = true
				break
			}
		}
		if exists_in_last == false {
			add_registrations[k3] = v3
		}
	}

	// Reconcile our adds and removes, if we find any that are "added" and also "removed", they cancel eachother out and can be removed from both.
	// This is normally a sign they already exist, and are being moved between 2 service names (extra host prefixes),
	// or moved from a service name to a prefix, or vice versa.
	add_registrations_results := make(Registrations)
	var remove_registrations_results []string
	// TODO: review
	/*
		for _, res_v1 := range add_host_prefixes {
			if stringInSlice(res_v1, remove_host_prefixes) == false {
				add_host_prefixes_results = append(add_host_prefixes_results, res_v1)
			}
		}
		for _, res_v2 := range remove_host_prefixes {
			if stringInSlice(res_v2, add_host_prefixes) == false {
				remove_host_prefixes_results = append(remove_host_prefixes_results, res_v2)
			}
		}
	*/

	return &add_registrations_results, &remove_registrations_results, nil
}
