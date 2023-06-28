package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	vers "github.com/hashicorp/go-version"
)

// Contains checks if a string slice contains a string.
//
// Returns true if present.
func Contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(e, a) {
			return true
		}
	}
	return false
}

// ContainsPosition is the same as Contains but returns the position of the string in the slice
func ContainsPosition(s []string, e string) (int, bool) {
	for i, a := range s {
		if strings.Contains(a, e) {
			return i, true
		}
	}
	return 0, false
}

// Remove will remove an item from a string slice and return the string slice
func Remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

// InvalidSemVerError is returned when a semantic version is invalid
type InvalidSemVerError struct {
	message string
}

func (e InvalidSemVerError) Error() string {
	return e.message
}

// CompareOSVer compares the reported OS string to the required.
// The version to check should be the first arg and the required
// version should be the second.
func CompareOSVer(a, b string) (bool, error) {
	uv, err := vers.NewVersion(a)
	if err != nil {
		return false, InvalidSemVerError{"Invalid semantic version"}
	}
	rv, err := vers.NewVersion(b)
	if err != nil {
		return false, InvalidSemVerError{"Invalid semantic version"}
	}
	// Comparison example. There is also GreaterThan, Equal, and just
	// a simple Compare that returns an int allowing easy >=, <=, etc.
	if uv.LessThan(rv) {
		return false, nil
	}

	return true, nil
}

// PossessiveForm returns the possessive form of a name. Or at least it tries to.
// For example, "John" becomes "John's" and "Silas" becomes "Silas'".
func PossessiveForm(name string) string {
	if len(name) == 0 {
		return name
	}

	lastChar := name[len(name)-1]
	possessiveForm := name + "'s"

	if lastChar == 's' {
		possessiveForm = name + "'"
	}

	return possessiveForm
}

func RemoveEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// RespToJson converts a response struct to json
func RespToJson(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(b)
}
