package env

import "github.com/zalando/go-keyring"

// RemoveKey removes a key from the keyring.
func RemoveKey(service, key string) error {
	return keyring.Delete(key, service)
}

// UpdateKey updates the value of a key in the keyring.
func UpdateKey(service, key, value string) error {
	return keyring.Set(key, value, service)
}
