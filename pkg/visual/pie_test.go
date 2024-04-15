package visual_test

import (
	"os"
	"testing"

	"github.com/johnmikee/cuebert/pkg/visual"
	"github.com/stretchr/testify/assert"
)

func TestPieChart(t *testing.T) {
	o := &visual.PieChartOption{
		ValueList: []float64{30, 50, 20},
		XAxis:     []string{"A", "B", "C"},
		Text:      "Pie Chart",
		Subtext:   "Subtitle",
		Query:     "query",
	}

	result, err := visual.PieChart(o)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// remove the generated file after the test
	os.RemoveAll(result)
}
