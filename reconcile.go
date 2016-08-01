package main

import ()

// Key = username@domain
// We use the <user> value from "sofia xmlstatus profile internal reg" to populate.
type Registrations map[string]string

// add_registrations Registrations, remove_registrations []string
func reconcileRegistrations(advertise_ip string, advertise_port string, last_active_registrations Registrations, current_active_registrations Registrations) (Registrations, []string, error) {
	var add_registrations Registrations
	var remove_registrations []string

	// TODO: we should only remove if the last registration was for this advertise IP.

	// First in one direction.
	for k1, _ := range last_active_registrations {
		name_exists_in_current := false
		for k2, _ := range current_active_registrations {
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
	for k3, v3 := range current_active_registrations {
		exists_in_last := false
		for k4, _ := range last_active_registrations {
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
	var add_registrations_results Registrations
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

	return add_registrations_results, remove_registrations_results, nil
}
