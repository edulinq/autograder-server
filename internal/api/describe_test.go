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
		customType reflect.Type
		expected   core.TypeDescription
	}{
		// Base types to alias (no JSON tags).
		{
			reflect.TypeOf((*string)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.BasicType,
				Alias:        "string",
			},
		},
		{
			reflect.TypeOf((*int)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.BasicType,
				Alias:        "int",
			},
		},
		{
			reflect.TypeOf((*int64)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.BasicType,
				Alias:        "int64",
			},
		},
		{
			reflect.TypeOf((*bool)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.BasicType,
				Alias:        "bool",
			},
		},

		// Simple wrapper types.
		{
			reflect.TypeOf((*stringWrapper)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.BasicType,
				Alias:        "string",
			},
		},
		{
			reflect.TypeOf((*core.MinServerRoleAdmin)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.BasicType,
				Alias:        "bool",
			},
		},

		// Simple maps and arrays.
		{
			reflect.TypeOf((*map[string]string)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.MapType,
				MapKeyType:   "string",
				MapValueType: "string",
			},
		},
		{
			reflect.TypeOf((*[]string)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory:     core.ArrayType,
				ArrayElementType: "string",
			},
		},

		// Wrapped maps and arrays.
		{
			reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.MapType,
				MapKeyType:   "string",
				MapValueType: "int",
			},
		},
		{
			reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory:     core.ArrayType,
				ArrayElementType: "bool",
			},
		},

		// Fields without JSON tags are ignored.
		{
			reflect.TypeOf((*simpleStruct)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.StructType,
				StructFields: map[string]string{},
			},
		},
		{
			reflect.TypeOf((*wrappedStruct)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.StructType,
				StructFields: map[string]string{},
			},
		},

		// Simple JSON tags.
		{
			reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.StructType,
				StructFields: map[string]string{
					"email":    "string",
					"job-code": "int",
				},
			},
		},

		// Hidden JSON tags.
		{
			reflect.TypeOf((*secureJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.StructType,
				StructFields: map[string]string{
					"first-name": "string",
					"last-name":  "string",
				},
			},
		},

		// Embedded fields.
		{
			reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.StructType,
				StructFields: map[string]string{
					"email":      "string",
					"job-code":   "int",
					"first-name": "string",
					"last-name":  "string",
				},
			},
		},

		// Complex fields.
		{
			reflect.TypeOf((*complexJSONStruct)(nil)).Elem(),
			core.TypeDescription{
				TypeCategory: core.StructType,
				StructFields: map[string]string{
					"coin-value": "map[string]{int}",
					"good-index": "[]{bool}",
					"personnel":  "email: string, first-name: string, job-code: int, last-name: string",
				},
			},
		},
	}

	for i, testCase := range testCases {
		typeMap := make(map[string]core.TypeDescription)

		actual := simplifyType(testCase.customType, typeMap)

		testCase.expected.TypeID = GetTypeID(testCase.customType)
		if !reflect.DeepEqual(testCase.expected, actual) {
			test.Errorf("Case %d: Unexpected type simplification. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(actual))
		}
	}
}
