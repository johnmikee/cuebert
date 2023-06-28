package auth

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/zalando/go-keyring"
)

/*
If we are developing locally default to using the keyring for safer storage of credentials

service: the name of the service the keys are being stored for.
to that end, its entirely up to you what you want to set this as. it can be globally set for all
team projects or switched on a project by project basis.
  - if you wish to set this the same for everything fill out the const value for svc

key: the name of the secret you are storing/getting
value: the value of the secret you are storing/getting
*/

var interactive bool = true

type Secret struct {
	Name  string
	Value string
}

type Secrets []Secret

const svc = "authAsaurusRex"

func notFound(errString string) bool {
	// keyring will return a specific message if the secret is not found.
	// if that is returned we know to add the secret. otherwise something
	// occurred that we need to handle differently.
	return errString == "secret not found in keyring"

}

func checkKey(service, key string) (string, bool) {
	// get password
	secret, err := keyring.Get(key, service)
	if err != nil {
		if notFound(err.Error()) {
			return "not-found", false
		}
		return "", false
	}

	return secret, true
}

func setKey(service, key string) error {
	log.Info().Bool("interactive", interactive).Msgf("setting key %s", key)
	value, err := SensitiveInputPrompt(
		&Prompter{
			Prompt:      BuildPrompt(key),
			Interactive: interactive,
			stub: &stubPasswordReader{
				Password:    key,
				ReturnError: false,
			},
		},
	)
	if err != nil {
		fmt.Printf("could not get the user-input for %s\n", key)
		return err
	}

	err = keyring.Set(key, service, value)
	if err != nil {
		fmt.Printf("error setting secret for %s in keyring.\n", key)
		return err
	}

	return nil
}

func getKey(service, key string) (string, bool) {
	secret, ok := checkKey(service, key)
	if !ok {
		if secret == "not-found" {
			err := setKey(service, key)
			if err != nil {
				return "", false
			}
			secret, _ = checkKey(service, key)
		}
	}
	return secret, true
}

// Optionally make it a map
func (s *Secrets) ToMap() map[string]string {
	result := make(map[string]string)

	for _, secret := range *s {
		result[secret.Name] = secret.Value
	}

	return result
}

func setInteractive(is bool) {
	interactive = is
}

// GetConfig will return a slice of key, values based on the args passed.
func GetConfig(service string, args ...string) *Secrets {
	secrets := Secrets{}
	if service == "" {
		service = svc
	}

	for _, a := range args {
		secret, ok := getKey(service, a)
		if ok {
			secrets = append(secrets, Secret{Name: a, Value: secret})
		}
	}

	return &secrets
}
