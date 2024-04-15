package parser

import (
	"reflect"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
)

func TestParser_Update(t *testing.T) {
	p := &Parser{
		Table: "my_table",
		Check: []CheckInfo{
			{Key: "column1", Fn: Prim{B: true}, Trimmed: "column1"},
			{Key: "column2", Fn: Prim{T: time.Now()}, Trimmed: "column2"},
			{Key: "column3", Fn: Prim{S: "value"}, Trimmed: "column3"},
			{Key: "column4", Fn: Prim{I: 42}, Trimmed: "column4"},
			{Key: "column5", Fn: Prim{I64: 64}, Trimmed: "column5"},
		},
		Index: "id",
		Val:   123,
	}

	base := sq.Update("my_table").Set("column0", "value0")

	sql, args, err := p.update(base)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedSQL := "UPDATE my_table SET column0 = ?, column1 = ?, column2 = ?, column3 = ?, column4 = ?, column5 = ? WHERE id = ?"
	expectedArgs := []interface{}{"value0", true, time.Now().Format(time.RFC3339), "value", 42, int64(64), 123}

	if sql != expectedSQL {
		t.Errorf("Unexpected SQL. Got: %s, Want: %s", sql, expectedSQL)
	}

	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Unexpected args. Got: %v, Want: %v", args, expectedArgs)
	}
}

func TestExtractCheckInfo(t *testing.T) {
	now := time.Now() // Capture the current time to use in your test case to minimize differences

	type TestStruct struct {
		Column1 string    `json:"column1"`
		Column2 int       `json:"column2"`
		Column3 int64     `json:"column3"`
		Column4 bool      `json:"column4"`
		Column5 time.Time `json:"column5"`
		Column6 string    `json:"-"`
	}

	tests := []struct {
		name     string
		input    interface{}
		expected []CheckInfo
	}{
		{
			name: "Test case 1",
			input: TestStruct{
				Column1: "value1",
				Column2: 42,
				Column3: 64,
				Column4: true,
				Column5: now,
				Column6: "value6",
			},
			expected: []CheckInfo{
				{Key: "column1", Fn: Prim{S: "value1"}, Trimmed: "Column1"},
				{Key: "column2", Fn: Prim{I: 42}, Trimmed: "Column2"},
				{Key: "column3", Fn: Prim{I64: 64}, Trimmed: "Column3"},
				{Key: "column4", Fn: Prim{B: true}, Trimmed: "Column4"},
				{Key: "column5", Fn: Prim{T: now}, Trimmed: "Column5"},
			},
		},
		// Add more test cases here if needed
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractCheckInfo(test.input)
			if !compareCheckInfoSlices(result, test.expected) {
				t.Errorf("Unexpected result. Got: %v, Want: %v", result, test.expected)
			}
		})
	}
}

func compareCheckInfoSlices(a, b []CheckInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Key != b[i].Key || a[i].Trimmed != b[i].Trimmed {
			return false
		}
		if !comparePrims(a[i].Fn, b[i].Fn) {
			return false
		}
	}
	return true
}

func comparePrims(a, b Prim) bool {
	if a.S != b.S || a.I != b.I || a.I64 != b.I64 || a.B != b.B {
		return false
	}
	// Compare times with a tolerance
	if !a.T.IsZero() && !b.T.IsZero() && a.T.Sub(b.T) > time.Second {
		return false
	}
	return true
}
