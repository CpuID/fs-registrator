package main

import (
	"reflect"
	"testing"
)

func TestRemoveSliceDuplicates(t *testing.T) {
	result := removeSliceDuplicates([]string{"aaaa", "bbbb", "aaaa", "cccc", "aaaa"})
	expected_result := []string{"aaaa", "bbbb", "cccc"}
	if reflect.DeepEqual(result, expected_result) != true {
		t.Errorf("Invalid result for removeSliceDuplicates(), expected %v, got %v", result, expected_result)
	}
}

func TestStringInSlice(t *testing.T) {
	result1 := stringInSlice("abc", []string{"abc", "def", "ghi"})
	if result1 != true {
		t.Errorf("Invalid result for stringInSlice(), expected true, got %t", result1)
	}
	result2 := stringInSlice("zzz", []string{"abc", "def", "ghi"})
	if result2 != false {
		t.Errorf("Invalid result for stringInSlice(), expected false, got %t", result2)
	}
}
