package util

import (
	"reflect"
	"testing"
)

type Unmarshalable struct {
	SomeFunc func() error
}

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
	Field3 string `json:"field3"`
}

func (this *Unmarshalable) String() string {
	return BaseString(this)
}

// Test that unmarshalable values do not infinite loop
// (since we may try to print them when writing errors).
func TestUnmarshalableStruct(test *testing.T) {
	value := &Unmarshalable{}

	_, err := ToJSON(value)
	if err == nil {
		test.Errorf("Unmarshalable object did not return error.")
	}
}

func TestToJsonMap(test *testing.T) {
	data := &TestStruct{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
	}

	expectedJSONMap := map[string]any{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	actualJSONMap, err := ToJSONMap(data)
	if err != nil {
		test.Errorf("Failed to convert data to a JSON map: '%v'.", err)
	}

	if !reflect.DeepEqual(actualJSONMap, expectedJSONMap) {
		test.Errorf("Unexpected JSON map. Expected: '%v', Actual: '%v'.",
			MustToJSONIndent(expectedJSONMap), MustToJSONIndent(actualJSONMap))
	}
}
