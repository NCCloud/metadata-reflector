package common

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustReturn[T any](t T, err error) T {
	Must(err)

	return t
}

func PointerTo[T any](t T) *T {
	return &t
}

func MapHasPrefix(prefix string, data map[string]string) bool {
	for key := range data {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// given a map, join its keys with comma as the separator
func MapKeysAsString(data map[string]string) string {
	return strings.Join(MapKeysAsSlice(data), ",")
}

// given 2 slices, check if the actual slice contains elements that are not present in expected
func ExcessiveElements(expected []string, actual []string) []string {
	var excessiveItems []string
	for _, element := range actual {
		if !slices.Contains(expected, element) {
			excessiveItems = append(excessiveItems, element)
		}
	}

	return excessiveItems
}

// get map keys as a slice
func MapKeysAsSlice(data map[string]string) []string {
	return slices.Collect(maps.Keys(data))
}

// check if any map key contains given substring
func MapContainsPartialKey(partialKey string, data map[string]string) bool {
	for key := range data {
		if strings.Contains(key, partialKey) {
			return true
		}
	}
	return false
}

func FindPartialKeys(partialKey string, data map[string]string) map[string]string {
	matchingData := make(map[string]string)
	for key, value := range data {
		if strings.Contains(key, partialKey) {
			matchingData[key] = value
		}
	}
	return matchingData
}

// given a pattern, make sure it will match the whole string
func ExactMatchRegex(pattern string) string {
	if !strings.HasPrefix(pattern, "^") {
		pattern = fmt.Sprintf("^%s", pattern)
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = fmt.Sprintf("%s$", pattern)
	}

	return pattern
}
