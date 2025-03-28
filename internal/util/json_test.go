package util

import (
	"reflect"
	"testing"
)

type Unmarshalable struct {
	SomeFunc func() error
}

type TestStruct struct {
	StringField string         `json:"string-field"`
	IntField    int            `json:"int-field"`
	FloatField  float64        `json:"float-field"`
	NullField   *string        `json:"null-field"`
	MapField    map[string]any `json:"map-field"`
	ListField   []string       `json:"list-field"`
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

func TestToJSONMap(test *testing.T) {
	data := &TestStruct{
		StringField: "string",
		IntField:    1,
		FloatField:  1.0,
		NullField:   nil,
		MapField:    map[string]any{"key": "value"},
		ListField:   []string{"string1", "string2", "string3"},
	}

	expectedJSONMap := map[string]any{
		"string-field": "string",
		"int-field":    1.0,
		"float-field":  1.0,
		"null-field":   nil,
		"map-field":    map[string]any{"key": "value"},
		"list-field":   []any{"string1", "string2", "string3"},
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
