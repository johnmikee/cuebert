package helpers

import (
	"fmt"
	"reflect"
	"testing"
)

type TestStruct struct {
	Field1 string `json:"field_1"`
	Field2 int    `json:"field_2"`
	Field3 bool   `json:"field_3,omitempty"`
	Field4 string
	Field5 string `json:"-"`
	Field6 string `json:"field_6,omitempty"`
	Field7 string `yaml:"field_7,omitempty"`
	Field8 string `yaml:"field_8"`
}

func TestGetStructKeys(t *testing.T) {
	expectedKeys := []string{"field_1", "field_2", "field_3", "field_6"}
	result := GetStructKeys(TestStruct{})

	if len(result) != len(expectedKeys) {
		t.Errorf("Test case failed: expected %d keys, got %d keys: %v", len(expectedKeys), len(result), result)
		return
	}

	for _, key := range expectedKeys {
		if !containsTag(result, key) {
			t.Errorf("Test case failed: expected key %q not found in any struct tag", key)
		}
	}
	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	user := User{
		ID:       1,
		Username: "john_doe",
		Email:    "john.doe@example.com",
	}

	expectedKeys = []string{"id", "username", "email"}
	keys := GetStructKeys(user)

	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("Test case failed: expected keys %v, got %v", expectedKeys, keys)
	}

}

func containsTag(tags []string, key string) bool {
	for _, tag := range tags {
		if tag == key {
			return true
		}
	}
	return false
}

func TestGetFieldName(t *testing.T) {
	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	// Test case: Valid struct and existing tag
	user := User{
		ID:       1,
		Username: "john_doe",
		Email:    "john.doe@example.com",
	}

	expectedFieldName := "Username"
	fieldName, err := GetFieldName("username", user)

	if err != nil {
		t.Errorf("Test case failed: unexpected error: %v", err)
	}

	if fieldName != expectedFieldName {
		t.Errorf("Test case failed: expected field name %s, got %s", expectedFieldName, fieldName)
	}

	// Test case: Invalid type (not a struct)
	var invalidStruct []int
	fmt.Printf("Type of invalidStruct: %T\n", invalidStruct) // Print the type for debugging
	invalidFieldName, err := GetFieldName("tag", invalidStruct)

	if invalidFieldName != "" {
		t.Errorf("Test case failed: expected empty field name, got %s", invalidFieldName)
	}

	expectedErrorMessage := "could not get field name"
	if err == nil || err.Error() != expectedErrorMessage {
		t.Errorf("Test case failed: expected error message '%s', got '%v'", expectedErrorMessage, err)
	}
}

func TestGetTagMapping(t *testing.T) {
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	expectedTagMapping := map[string]interface{}{
		"name":    "Name",
		"age":     "Age",
		"address": "Address",
	}

	tagMapping := GetTagMapping(reflect.TypeOf(Person{}))

	if !reflect.DeepEqual(tagMapping, expectedTagMapping) {
		t.Errorf("Test case failed: expected tag mapping %v, got %v", expectedTagMapping, tagMapping)
	}
}

func TestBuildFieldsByTagMap(t *testing.T) {
	type User struct {
		ID       int    `json:"-"`
		Username string `json:"username"`
		Email    string `json:"email,omitempty"`
		Password string `json:",omitempty"`
	}

	user := User{
		ID:       1,
		Username: "john_doe",
		Email:    "john.doe@example.com",
		Password: "secret",
	}

	buildFieldsByTagMap(user)

	expectedFields := map[string]map[string]string{
		keyType: {
			"username": "Username",
			"email":    "Email",
		},
	}

	if !reflect.DeepEqual(fieldsByTag[reflect.TypeOf(user)], expectedFields) {
		t.Errorf("Test case failed: expected fields map %v, got %v", expectedFields, fieldsByTag[reflect.TypeOf(user)])
	}
}
