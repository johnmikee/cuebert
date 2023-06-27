package visual

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildOut(t *testing.T) {
	tmpPath := filepath.Join(os.TempDir(), "visual")

	// Clean up the temporary directory after the test finishes
	defer func() {
		err := os.RemoveAll(tmpPath)
		assert.NoError(t, err)
	}()

	input := "test-input"
	visType := Pie
	buf := []byte{1, 2, 3, 4}

	result, err := buildOut(input, nil, visType, buf)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify that the file exists in the temporary directory
	fileExists := fileExists(result)
	assert.True(t, fileExists)
}

func TestBuildOutErr(t *testing.T) {
	// Create a temporary directory that we know will fail
	// TODO: this is only for macOS via a SIP-protected directory.
	// We should find a way to make this work on other platforms.
	tmpBadPath := "/Library/Updates/new"
	// Clean up the temporary directory after the test finishes
	defer func() {
		err := os.RemoveAll(tmpBadPath)
		assert.NoError(t, err)
	}()

	input := "test-input"
	visType := Pie
	buf := []byte{1, 2, 3, 4}

	// This should cause an error
	result, err := buildOut(input, &tmpBadPath, visType, buf)

	assert.Error(t, err)
	assert.NotEmpty(t, result)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Test_buildOut(t *testing.T) {
	type args struct {
		input   string
		output  *string
		visType VisualType
		buf     []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildOut(tt.args.input, tt.args.output, tt.args.visType, tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildOut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildOut() = %v, want %v", got, tt.want)
			}
		})
	}
}
