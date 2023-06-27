package helpers

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCompareOSVer(t *testing.T) {
	testCases := []struct {
		a              string
		b              string
		expectedResult bool
		expectedError  error
	}{
		{"1.2.3", "1.2.4", false, nil},
		{"1.2.3", "1.2.3", true, nil},
		{"1.2.4", "1.2.3", true, nil},
		{"1.2.3", "2.0.0", false, nil},
		{"invalid", "1.2.3", false, InvalidSemVerError{"Invalid semantic version"}},
		{"9.5.6", "4fg4:.3", false, InvalidSemVerError{"Invalid semantic version"}},
	}

	for _, testCase := range testCases {
		result, err := CompareOSVer(testCase.a, testCase.b)

		if result != testCase.expectedResult {
			t.Errorf("Test case failed: a=%s, b=%s, expected result=%v, got=%v", testCase.a, testCase.b, testCase.expectedResult, result)
		}

		if !errors.Is(err, testCase.expectedError) {
			t.Errorf("Test case failed: a=%s, b=%s, expected error=%v, got=%v", testCase.a, testCase.b, testCase.expectedError, err)
		}
	}
}

func TestContains(t *testing.T) {
	slice := []string{
		"Apple I",
		"Apple II",
		"Macintosh 128K",
		"Macintosh SE",
		"Macintosh II",
	}

	exists := "Macintosh 128K"
	result := Contains(slice, exists)
	if !result {
		t.Errorf("Test case failed: expected %s to be found in slice", exists)
	}

	// Test case: element does not exist in the slice
	notExists := "Macbook Pro"
	result = Contains(slice, notExists)
	if result {
		t.Errorf("Test case failed: unexpected %s found in slice", notExists)
	}
}

func TestContainsPosition(t *testing.T) {
	slice := []string{"munki", "autopkg", "santa", "osquery", "salt"}

	// Test case: element exists in the slice
	exists := "salt"
	expectedIndex := 4
	resultIndex, result := ContainsPosition(slice, exists)
	if !result {
		t.Errorf("Test case failed: expected %s to be found in slice", exists)
	}
	if resultIndex != expectedIndex {
		t.Errorf("Test case failed: expected index=%d, got index=%d", expectedIndex, resultIndex)
	}

	// Test case: element does not exist in the slice
	notExists := "jamf"
	expectedIndex = 0
	resultIndex, result = ContainsPosition(slice, notExists)
	if result {
		t.Errorf("Test case failed: unexpected %s found in slice", notExists)
	}
	if resultIndex != expectedIndex {
		t.Errorf("Test case failed: expected index=%d, got index=%d", expectedIndex, resultIndex)
	}
}

func TestPossessiveForm(t *testing.T) {
	testCases := []struct {
		name           string
		expectedResult string
	}{
		{"John", "John's"},
		{"Mary", "Mary's"},
		{"Chris", "Chris'"},
		{"James", "James'"},
		{"Lucy", "Lucy's"},
		{"", ""},
	}

	for _, testCase := range testCases {
		result := PossessiveForm(testCase.name)
		if result != testCase.expectedResult {
			t.Errorf("Test case failed: name=%s, expected=%s, got=%s", testCase.name, testCase.expectedResult, result)
		}
	}
}

func TestRemove(t *testing.T) {
	slice := []string{"Gopher", "banana", "macOS", "date", "Golang"}
	index := 2 // Index of "macOS" to be removed
	expectedResult := []string{"Gopher", "banana", "date", "Golang"}

	result := Remove(slice, index)

	if len(result) != len(expectedResult) {
		t.Errorf("Test case failed: expected length=%d, got=%d", len(expectedResult), len(result))
	}

	for i := range result {
		if result[i] != expectedResult[i] {
			t.Errorf("Test case failed: expected %s, got %s", expectedResult[i], result[i])
		}
	}
}

func TestInvalidSemVerError_Error(t *testing.T) {
	err := InvalidSemVerError{"Invalid semantic version"}
	expectedResult := "Invalid semantic version"
	result := err.Error()

	if result != expectedResult {
		t.Errorf("Test case failed: expected %s, got %s", expectedResult, result)
	}
}

func TestRemoveEmpty(t *testing.T) {
	slice := []string{"iOS 12", "", "macOS Catalina", "", "", "watchOS 7", "tvOS 14", "", "macOS Big Sur"}
	expectedResult := []string{"iOS 12", "macOS Catalina", "watchOS 7", "tvOS 14", "macOS Big Sur"}

	result := RemoveEmpty(slice)

	if len(result) != len(expectedResult) {
		t.Errorf("Test case failed: expected length=%d, got=%d", len(expectedResult), len(result))
	}

	for i := range result {
		if result[i] != expectedResult[i] {
			t.Errorf("Test case failed: expected %s, got %s", expectedResult[i], result[i])
		}
	}
}

func TestRespToJson(t *testing.T) {
	type Gopher struct {
		Name     string
		Message  string
		Birthday string
	}

	gopher := Gopher{
		Name:     "Glen(da) Jr.",
		Message:  "Look, that rabbit's got a vicious streak a mile wide! It's a killer!",
		Birthday: "Tuesday, November 10, 2009",
	}

	expectedResult := `{"Name":"Glen(da) Jr.","Message":"Look, that rabbit's got a vicious streak a mile wide! It's a killer!","Birthday":"Tuesday, November 10, 2009"}`

	result := RespToJson(gopher)

	if result != expectedResult {
		t.Errorf("Test case failed: expected %s, got %s", expectedResult, result)
	}

	// Test case: Error handling
	invalidStruct := make(chan int) // Invalid struct that cannot be marshaled to JSON

	expectedError := errors.New("unexpected end of JSON input")
	result = RespToJson(invalidStruct)

	if result != "" {
		t.Errorf("Test case failed: expected empty string, got %s", result)
	}

	var unmarshalResult Gopher
	err := json.Unmarshal([]byte(result), &unmarshalResult)

	if err == nil {
		t.Errorf("Test case failed: expected error, got nil")
	} else if err.Error() != expectedError.Error() {
		t.Errorf("Test case failed: expected error=%v, got %v", expectedError, err.Error())
	}

}
