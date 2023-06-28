package visual

import (
	"fmt"
	"os"
	"path/filepath"
)

type VisualType string

const (
	Pie      VisualType = "pie"
	BarGraph VisualType = "bar"
)

var tmpPath = filepath.Join(os.TempDir(), "visual")

func buildOut(input string, output *string, visType VisualType, buf []byte) (string, error) {
	var createPath string
	if output != nil {
		createPath = *output
	} else {
		createPath = tmpPath
	}

	err := os.MkdirAll(createPath, 0700)
	if err != nil {
		return tmpPath, err
	}

	file := filepath.Join(tmpPath, fmt.Sprintf("%s-%s.png", input, visType))
	err = os.WriteFile(file, buf, 0600)

	return file, err
}
