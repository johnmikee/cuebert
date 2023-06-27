package client

import (
	"testing"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/mdm/jamf"
	"github.com/johnmikee/cuebert/mdm/kandji"
)

func TestNewJamf(t *testing.T) {
	// Create a sample MDM configuration
	config := mdm.Config{
		Domain: "example.com",
		MDM:    mdm.Jamf,
		// Other configuration fields
	}

	// Create a sample MDM instance
	mdmInstance := MDM{
		MDM:    mdm.Jamf,
		Config: config,
	}

	// Call the New function to create the MDM provider
	provider := New(&mdmInstance)

	// Check the type of the returned provider
	switch provider.(type) {
	case *jamf.Client:
		// The provider is of type *jamf.Client, which is expected for Jamf MDM
		t.Log("Jamf provider created successfully")
	case *kandji.Config:
		t.Error("Expected Jamf provider, but got Kandji provider")
	default:
		t.Error("Unknown provider type")
	}
}

func TestNewKandji(t *testing.T) {
	// Create a sample MDM configuration
	config := mdm.Config{
		Domain: "example.com",
		MDM:    mdm.Kandji,
		// Other configuration fields
	}

	// Create a sample MDM instance
	mdmInstance := MDM{
		MDM:    mdm.Kandji,
		Config: config,
	}

	// Call the New function to create the MDM provider
	provider := New(&mdmInstance)

	// Check the type of the returned provider
	switch provider.(type) {
	case *jamf.Client:
		t.Error("Expected Kandji provider, but got Jamf provider")
	case *kandji.Config:
		// The provider is of type *kandji.Config, which is expected for Kandji MDM
		t.Log("Kandji provider created successfully")
	default:
		t.Error("Unknown provider type")
	}
}
