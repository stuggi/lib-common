/*
Copyright 2020 Red Hat

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"encoding/json"
	"strings"
)

// GetOr returns the value of m[key] if it exists, fallback otherwise.
// As a special case, it also returns fallback if the value of m[key] is
// the empty string
func GetOr(m map[string]interface{}, key, fallback string) interface{} {
	val, ok := m[key]
	if !ok {
		return fallback
	}

	s, ok := val.(string)
	if ok && s == "" {
		return fallback
	}

	return val
}

// IsSet returns the value of m[key] if key exists, otherwise false
// Different from getOr because it will return zero values.
func IsSet(m map[string]interface{}, key string) interface{} {
	val, ok := m[key]
	if !ok {
		return false
	}
	return val
}

// IsJSON check if string is json format
func IsJSON(s string) error {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js)
}

// RemoveIndex - remove int from slice
func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

// RemoveValue removes all occurrences of a specific value from a string slice.
// It returns the modified (shorter) slice.
func RemoveValue(s []string, valueToRemove string) []string {
	// Use a new index 'n' to track the position of the next element to keep.
	n := 0
	for _, v := range s {
		// If the current value 'v' is not the one to remove,
		// keep it by moving it to the front of the slice.
		if v != valueToRemove {
			s[n] = v
			n++
		}
	}
	// Truncate the slice to its new length 'n'.
	// The underlying array is the same, but the slice is now shorter.
	return s[:n]
}

// RemoveValues removes all occurrences of strings that are present in the 'valuesToRemove' slice.
// It uses a map for efficient lookups and modifies the original slice in place.
func RemoveValues(s []string, valuesToRemove []string) []string {
	// 1. Create a map from the 'valuesToRemove' slice for fast lookups.
	// We only care about the keys.
	toRemoveSet := make(map[string]struct{}, len(valuesToRemove))
	for _, v := range valuesToRemove {
		toRemoveSet[v] = struct{}{}
	}

	// 2.  Use a new index 'n' to track the position of the next element to keep.
	n := 0
	for _, v := range s {
		// Check if the current value 'v' is in our removal set.
		// If the current value 'v' is not the one to remove,
		// keep it by moving it to the front of the slice.
		if _, found := toRemoveSet[v]; !found {
			s[n] = v
			n++
		}
	}

	// 3. Truncate the slice to its new length 'n'.
	// The underlying array is the same, but the slice is now shorter.
	return s[:n]
}

// KeepOnlyValues modifies a slice in place to keep only the elements
// that are present in the 'valuesToKeep' slice.
func KeepOnlyValues(s []string, valuesToKeep []string) []string {
	// 1. Create a map from the 'valuesToKeep' slice for fast lookups.
	// We only care about the keys.
	toKeepSet := make(map[string]struct{}, len(valuesToKeep))
	for _, v := range valuesToKeep {
		toKeepSet[v] = struct{}{}
	}

	// 2.  Use a new index 'n' to track the position of the next element to keep.
	n := 0
	for _, v := range s {
		// Check if the current value 'v' is in our whitelist.
		// If it is found we keep it by moving it to the front of the slice.
		if _, found := toKeepSet[v]; found {
			s[n] = v
			n++
		}
	}

	// 3. Truncate the slice to its new length 'n'.
	// The underlying array is the same, but the slice is now shorter.
	return s[:n]
}

// KeepIfContainsSubstring modifies a slice in place to keep only the elements
// that contain any of the provided substrings.
func KeepIfContainsSubstring(s []string, substrings []string) []string {
	// 'n' is the index to track the position of the next element to keep.
	n := 0
	for _, item := range s {
		// Assume we don't keep the item until we find a matching substring.
		shouldKeep := false
		// Check against every required substring.
		for _, sub := range substrings {
			if sub == "" {
				continue
			}

			if strings.Contains(item, sub) {
				shouldKeep = true
				break
			}
		}

		// If the flag was set, we keep it by moving it to the front of the slice.
		if shouldKeep {
			s[n] = item
			n++
		}
	}

	// Truncate the slice to its new length 'n'.
	// The underlying array is the same, but the slice is now shorter.
	return s[:n]
}
