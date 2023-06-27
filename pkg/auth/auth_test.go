package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

const testSvc = "authAsaurusRex"

// teardown deletes the keys from the keyring
func teardown(keys []string) {
	for _, key := range keys {
		keyring.Delete(key, svc)
	}
}

// TestSetSimpleKeys tests setting simple keys in the keyring
func TestSetSimpleKeys(t *testing.T) {
	testKeys := []string{"test1", "test2", "test3"}
	for _, key := range testKeys {
		err := keyring.Set(testSvc, key, "test")
		if err != nil {
			t.Fatal(err)
		}
	}
	teardown(testKeys)
}

// TestGetSetSimpleKeys tests setting and getting simple keys in the keyring
func TestGetSetSimpleKeys(t *testing.T) {
	testInput := make(map[string]string)
	testInput["key1"] = "value1"
	testInput["key2"] = "value2"
	testInput["key3"] = "value3"
	testInput["key4"] = "value4"
	testInput["key5"] = "value5"

	for k, v := range testInput {
		err := keyring.Set(testSvc, k, v)
		if err != nil {
			t.Fatal(err)
		}

		secret, err := keyring.Get(testSvc, k)
		if err != nil {
			t.Fatal(err)
		}

		if secret != v {
			t.Fatalf("expected %s, got %s", "test", secret)
		}
	}

	teardown([]string{"key1", "key2", "key3", "key4", "key5"})
}

var goIntro = `
Introduction

Go is a new language. Although it borrows ideas from existing languages, it has unusual properties that make effective Go programs different in character from programs written in its relatives. A straightforward translation of a C++ or Java program into Go is unlikely to produce a satisfactory result—Java programs are written in Java, not Go. On the other hand, thinking about the problem from a Go perspective could produce a successful but quite different program. In other words, to write Go well, it's important to understand its properties and idioms. It's also important to know the established conventions for programming in Go, such as naming, formatting, program construction, and so on, so that programs you write will be easy for other Go programmers to understand.
`

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	buffer := make([]byte, length/2) // Since hex encoding requires 2 characters per byte

	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	randomString := hex.EncodeToString(buffer)
	return randomString, nil
}

// generateRandomStringArray generates an array of random strings
func generateRandomStringArray(arrayLength, stringLength int) ([]string, error) {
	randomStrings := make([]string, arrayLength)

	for i := 0; i < arrayLength; i++ {
		randomString, err := generateRandomString(stringLength)
		if err != nil {
			return nil, err
		}

		randomStrings[i] = randomString
	}

	return randomStrings, nil
}

// TestGetSetLongKeys tests setting and getting long keys in the keyring
func TestGetSetLongKeys(t *testing.T) {
	arrayLength := 10  // Length of the string array
	stringLength := 16 // Length of each random string (in bytes)

	randomStringArray, err := generateRandomStringArray(arrayLength, stringLength)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	testInput := make(map[string]string)
	testInput["key1"] = goIntro
	testInput["key2"] = base64.StdEncoding.EncodeToString([]byte(goIntro))
	testInput["key3"] = "value3"
	testInput["key4"] = strings.Join(randomStringArray, ",")
	testInput["key5"] = "???????????????????????????????????"

	for k, v := range testInput {
		err := keyring.Set(testSvc, k, v)
		if err != nil {
			t.Fatal(err)
		}

		secret, err := keyring.Get(testSvc, k)
		if err != nil {
			t.Fatal(err)
		}

		if secret != v {
			t.Fatalf("expected %s, got %s", "test", secret)
		}
	}

	teardown([]string{"key1", "key2", "key3", "key4", "key5"})
}

// TestCheckKeyNotFound tests the scenario when a key is not found in the keyring
func TestCheckKeyNotFound(t *testing.T) {
	arrayLength := 4   // Length of the string array
	stringLength := 12 // Length of each random string (in bytes)

	randomStringArray, err := generateRandomStringArray(arrayLength, stringLength)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	service := "testService"
	key := strings.Join(randomStringArray, ",")

	secret, found := checkKey(service, key)

	if found {
		t.Errorf("Expected key not to be found, but it was found. Secret: %s", secret)
	}
}

// TestCheckKeyFound tests the scenario when a key is found in the keyring
func TestCheckKeyFound(t *testing.T) {
	service := "testService"
	key := "testKey"
	value := "testValue"

	err := keyring.Set(key, service, value)
	if err != nil {
		t.Fatalf("Failed to set key. Error: %s", err.Error())
	}

	secret, found := checkKey(service, key)

	if !found {
		t.Error("Expected key to be found, but it was not found.")
	}

	if secret != value {
		t.Errorf("Expected secret value %s, but got %s", value, secret)
	}

	teardown([]string{key})
}

// TestGetKeyNotFound tests the scenario when a key is not found in the keyring
func TestGetKeyNotFound(t *testing.T) {
	arrayLength := 4   // Length of the string array
	stringLength := 12 // Length of each random string (in bytes)

	randomStringArray, err := generateRandomStringArray(arrayLength, stringLength)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	service := "testService"
	key := strings.Join(randomStringArray, ",")
	setInteractive(false)
	secret, found := getKey(service, key)

	if !found {
		t.Errorf("Expected key to be set when not found. Secret: %s", secret)
	}
}

// TestGetKeyFound tests the scenario when a key is found in the keyring
func TestGetKeyFound(t *testing.T) {
	service := "testService"
	key := "testKey"
	value := "testValue"

	err := keyring.Set(key, service, value)
	if err != nil {
		t.Fatalf("Failed to set key. Error: %s", err.Error())
	}

	secret, found := getKey(service, key)

	if !found {
		t.Error("Expected key to be found, but it was not found.")
	}

	if secret != value {
		t.Errorf("Expected secret value %s, but got %s", value, secret)
	}

	teardown([]string{key})
}

// TestToMap tests the ToMap method of the Secrets type
func TestToMap(t *testing.T) {
	secrets := Secrets{
		{Name: "key1", Value: "value1"},
		{Name: "key2", Value: "value2"},
	}
	expectedMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	result := secrets.ToMap()
	if len(result) != len(expectedMap) {
		t.Errorf("expected map length=%d, got length=%d", len(expectedMap), len(result))
	}
	for k, v := range expectedMap {
		if result[k] != v {
			t.Errorf("expected value=%s for key=%s, got value=%s", v, k, result[k])
		}
	}
}

// TestCheckKey tests the checkKey function
func TestCheckKey(t *testing.T) {
	// Test when secret is found
	err := keyring.Set(testSvc, "key1", "test")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	secret, ok := checkKey("key1", testSvc)
	if !ok {
		t.Errorf("expected ok=true, got ok=false")
	}
	if secret != "test" {
		t.Errorf("expected secret=test, got secret=%s", secret)
	}

	// Test when secret is not found
	secret, ok = checkKey("nonexistent", testSvc)
	if ok {
		t.Errorf("expected ok=false, got ok=true")
	}
	if secret != "not-found" {
		t.Errorf("expected secret='not-found', got secret=%s", secret)
	}
}

// TestGetConfig tests the GetConfig function
func TestGetConfig(t *testing.T) {
	err := keyring.Set(testSvc, "key1", "test")
	if err != nil {
		t.Fatal(err)
	}
	setInteractive(false)

	// Test case 1: Get existing secrets
	secrets := GetConfig(testSvc, "key1", "key2")
	if len(*secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(*secrets))
	}

	// Verify that the secrets are retrieved correctly
	expectedSecrets := map[string]string{
		"key1": "key1",
		"key2": "key2",
	}
	for _, secret := range *secrets {
		fmt.Println(secret)
		expectedValue, ok := expectedSecrets[secret.Name]
		if !ok {
			t.Errorf("Unexpected secret retrieved: %s", secret.Name)
		}
		if secret.Value != expectedValue {
			t.Errorf("Incorrect value for secret %s. Expected %s, got %s", secret.Name, expectedValue, secret.Value)
		}
	}

	// Test case 2: Get non-existent secret
	_, err = keyring.Get(testSvc, "THISkey5")
	if err != nil {
		assert.Equal(t, err.Error(), "secret not found in keyring")
	}

	// Test case 3: Empty service name
	secrets = GetConfig("", "key1", "key2")
	if len(*secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(*secrets))
	}

}
