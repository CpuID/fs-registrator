package main

import ()

// Credit - http://stackoverflow.com/questions/15323767/does-golang-have-if-x-in-construct-similar-to-python
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Credit - https://groups.google.com/forum/#!topic/golang-nuts/-pqkICuokio
func removeSliceDuplicates(data []string) []string {
	length := len(data) - 1
	for i := 0; i < length; i++ {
		for j := i + 1; j <= length; j++ {
			if data[i] == data[j] {
				data[j] = data[length]
				data = data[0:length]
				length--
				j--
			}
		}
	}
	return data
}
