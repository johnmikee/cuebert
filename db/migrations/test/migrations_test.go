package migrations_test

import (
	b64 "encoding/base64"
	"strings"
	"testing"

	"github.com/johnmikee/cuebert/db/migrations"
)

var connStr string = "cG9zdGdyZXM6Ly9jdWU6Y3VlQGxvY2FsaG9zdDo1NDMyL2N1ZT9zc2xtb2RlPWRpc2FibGUK"

func bdc() string {
	sDec, _ := b64.StdEncoding.DecodeString(connStr)
	return strings.TrimSpace(string(sDec))
}

func TestMake(t *testing.T) {
	migrationFiles := "../sql"
	err := migrations.Make(bdc(), migrationFiles)
	if err != nil {
		t.Fatalf("Failed to execute Make: %v", err)
	}
}

func TestMakeEmbed(t *testing.T) {
	err := migrations.MakeEmbed(bdc())
	if err != nil {
		t.Fatalf("Failed to execute MakeEmbed: %v", err)
	}
}

func TestMakeWithInvalidConnectionString(t *testing.T) {
	connStr := "invalid-connection-string"
	migrationFiles := "../sql"
	err := migrations.Make(connStr, migrationFiles)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
}
