package api

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestDescribeFull(test *testing.T) {
	path, err := util.GetAPIDescriptionFilepath()
	if err != nil {
		test.Fatalf("Unable to get the API description filepath: '%v'.", err)
	}

	var expectedDescriptions core.APIDescription
	err = util.JSONFromFile(path, &expectedDescriptions)
	if err != nil {
		test.Fatalf("Failed to load api.json: '%v'.", err)
	}

	descriptions, err := Describe(*GetRoutes())
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	if !reflect.DeepEqual(&expectedDescriptions, descriptions) {
		test.Fatalf("Unexpected API Descriptions. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expectedDescriptions), util.MustToJSONIndent(descriptions))
	}
}

func TestDescribeEmpty(test *testing.T) {
	descriptions, err := Describe(*GetRoutes())
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	for endpoint, description := range descriptions.Endpoints {
		if description.Description == "" {
			test.Errorf("Describe found an empty description. Endpoint: '%s'.", endpoint)
			continue
		}
	}
}

func TestDescribeEmptyRoutes(test *testing.T) {
	routes := []core.Route{}
	description, err := Describe(routes)
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	if len(description.Endpoints) != 0 {
		test.Errorf("Unexpected number of endpoints. Expected: '0', actual: '%d'.", len(description.Endpoints))
	}
}

type stringWrapper string
type simpleMapWrapper map[string]int
type simpleArrayWrapper []bool

type simpleStruct struct {
	BaseString string
	BaseInt    int
	BaseBool   bool
}

type wrappedStruct struct {
	WrappedString stringWrapper
	WrappedMap    simpleMapWrapper
	WrappedArray  simpleArrayWrapper
}

type simpleJSONStruct struct {
	Email   string `json:"email"`
	JobCode int    `json:"job-code"`
}

type secureJSONStruct struct {
	FirstName string `json:"first-name"`
	LastName  string `json:"last-name"`
	Pay       int    `json:"-"`
}

type embeddedJSONStruct struct {
	simpleJSONStruct
	secureJSONStruct
}

type complexJSONStruct struct {
	CoinValue simpleMapWrapper   `json:"coin-value"`
	GoodIndex simpleArrayWrapper `json:"good-index"`
	Personnel embeddedJSONStruct `json:"personnel"`
}

func TestSimplifyType(test *testing.T) {
	testCases := []struct {
		customType      reflect.Type
		expectedDesc    core.TypeDescription
		expectedTypeMap map[string]core.TypeDescription
	}{
		// Base types to alias (no JSON tags).
		{
			reflect.TypeOf((*string)(nil)).Elem(),
			core.TypeDescription{
				Category: core.BasicType,
				Alias:    "string",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*int)(nil)).Elem(),
			core.TypeDescription{
				Category: core.BasicType,
				Alias:    "int",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*int64)(nil)).Elem(),
			core.TypeDescription{
				Category: core.BasicType,
				Alias:    "int64",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*bool)(nil)).Elem(),
			core.TypeDescription{
				Category: core.BasicType,
				Alias:    "bool",
			},
			map[string]core.TypeDescription{},
		},

		// Simple wrapper types.
		{
			reflect.TypeOf((*stringWrapper)(nil)).Elem(),
			core.TypeDescription{
				Category: core.BasicType,
				Alias:    "string",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*core.MinServerRoleAdmin)(nil)).Elem(),
			core.TypeDescription{
				Category: core.BasicType,
				Alias:    "bool",
			},
			map[string]core.TypeDescription{},
		},

		// Simple maps and arrays.
		{
			reflect.TypeOf((*map[string]string)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]core.TypeDescription{
				"map[string]string": core.TypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "string",
				},
			},
		},
		{
			reflect.TypeOf((*[]string)(nil)).Elem(),
			core.TypeDescription{
				Category:    core.ArrayType,
				ElementType: "string",
			},
			map[string]core.TypeDescription{
				"[]string": core.TypeDescription{
					Category:    core.ArrayType,
					ElementType: "string",
				},
			},
		},

		// Wrapped maps and arrays.
		{
			reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "int",
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem()): core.TypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
		{
			reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(),
			core.TypeDescription{
				Category:    core.ArrayType,
				ElementType: "bool",
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem()): core.TypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
			},
		},

		// Fields without JSON tags are ignored.
		{
			reflect.TypeOf((*simpleStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields:   nil,
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*simpleStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields:   nil,
				},
			},
		},
		{
			reflect.TypeOf((*wrappedStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields:   nil,
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*wrappedStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields:   nil,
				},
			},
		},

		// Simple JSON tags.
		{
			reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields: map[string]string{
					"email":    "string",
					"job-code": "int",
				},
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":    "string",
						"job-code": "int",
					},
				},
			},
		},

		// Hidden JSON tags (-).
		{
			reflect.TypeOf((*secureJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields: map[string]string{
					"first-name": "string",
					"last-name":  "string",
				},
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"first-name": "string",
						"last-name":  "string",
					},
				},
			},
		},

		// Embedded fields.
		{
			reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields: map[string]string{
					"email":      "string",
					"job-code":   "int",
					"first-name": "string",
					"last-name":  "string",
				},
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":      "string",
						"first-name": "string",
						"job-code":   "int",
						"last-name":  "string",
					},
				},
				GetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":    "string",
						"job-code": "int",
					},
				},
				GetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"first-name": "string",
						"last-name":  "string",
					},
				},
			},
		},

		// Complex fields.
		{
			reflect.TypeOf((*complexJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields: map[string]string{
					"coin-value": "github.com/edulinq/autograder/internal/api/api.simpleMapWrapper",
					"good-index": "github.com/edulinq/autograder/internal/api/api.simpleArrayWrapper",
					"personnel":  "github.com/edulinq/autograder/internal/api/api.embeddedJSONStruct",
				},
			},
			map[string]core.TypeDescription{
				GetTypeID(reflect.TypeOf((*complexJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"coin-value": "github.com/edulinq/autograder/internal/api/api.simpleMapWrapper",
						"good-index": "github.com/edulinq/autograder/internal/api/api.simpleArrayWrapper",
						"personnel":  "github.com/edulinq/autograder/internal/api/api.embeddedJSONStruct",
					},
				},
				GetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":      "string",
						"first-name": "string",
						"job-code":   "int",
						"last-name":  "string",
					},
				},
				GetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"first-name": "string",
						"last-name":  "string",
					},
				},
				GetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem()): core.TypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
				GetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":    "string",
						"job-code": "int",
					},
				},
				GetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem()): core.TypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
	}

	for i, testCase := range testCases {
		typeMap := make(map[string]core.TypeDescription)

		actual, _ := simplifyType(testCase.customType, typeMap)

		if !reflect.DeepEqual(testCase.expectedDesc, actual) {
			test.Errorf("Case %d: Unexpected type simplification. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedDesc), util.MustToJSONIndent(actual))
		}

		if !reflect.DeepEqual(testCase.expectedTypeMap, typeMap) {
			test.Errorf("Case %d: Unexpected type map. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedTypeMap), util.MustToJSONIndent(typeMap))
		}
	}
}
