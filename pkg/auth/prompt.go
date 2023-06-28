package auth

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/term"
)

/*
inspired by https://petersouter.xyz/testing-and-mocking-stdin-in-golang/
mocking user input was a bit tricky, but this article saved me oodles of time.
*/

// PasswordReader returns password read from a reader
type PasswordReader interface {
	ReadPassword() (string, error)
}

// StdInPasswordReader default stdin password reader
type StdInPasswordReader struct {
}

// ReadPassword reads password from stdin
func (pr StdInPasswordReader) ReadPassword() (string, error) {
	pwd, error := term.ReadPassword(int(syscall.Stdin))
	return string(pwd), error
}

func BuildPrompt(text string) string {
	return fmt.Sprint("Please enter the value for ", text)
}

func readPassword(pr PasswordReader) (string, error) {
	pwd, err := pr.ReadPassword()
	if err != nil {
		return "", err
	}
	if len(pwd) == 0 {
		return "", errors.New("empty password provided")
	}
	return pwd, nil
}

func run(pr PasswordReader) (string, error) {
	pwd, err := readPassword(pr)
	if err != nil {
		return "", err
	}

	return string(pwd), nil
}

type stubPasswordReader struct {
	Password    string
	ReturnError bool
}

func (pr stubPasswordReader) ReadPassword() (string, error) {
	if pr.ReturnError {
		return "", errors.New("stubbed error")
	}
	return pr.Password, nil
}

type Prompter struct {
	Prompt      string
	Interactive bool
	stub        *stubPasswordReader
}

func SensitiveInputPrompt(p *Prompter) (string, error) {
	var (
		pwdr     PasswordReader
		password string
		err      error
	)

	if p.Interactive {
		fmt.Printf("\x1b[32m%s: \x1b[0m\n", p.Prompt)
		password, err = run(pwdr)
	} else {
		pr := p.stub
		password, err = run(pr)
	}

	if err != nil {
		return "", fmt.Errorf("error reading password: %w", err)
	}

	return password, nil
}
