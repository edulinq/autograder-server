package api

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

type stringWrapper string
type simpleMapWrapper map[string]int
type simpleArrayWrapper []bool

type simplePointerWrapper *string

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

type complexPointerStruct struct {
	CoinValue *simpleMapWrapper   `json:"coin-value"`
	GoodIndex *simpleArrayWrapper `json:"good-index"`
	Personnel *embeddedJSONStruct `json:"personnel"`
}

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

	actualDescriptions, err := Describe(*GetRoutes())
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	if !reflect.DeepEqual(&expectedDescriptions, actualDescriptions) {
		// If not deep equal, check for JSON equality.
		descriptionString := util.MustToJSON(actualDescriptions)
		var descriptions core.APIDescription
		util.MustJSONFromString(descriptionString, &descriptions)

		if !reflect.DeepEqual(expectedDescriptions, descriptions) {
			message := "Unexpected API Descriptions. "

			for endpoint, expectedDesc := range expectedDescriptions.Endpoints {
				actualDesc, ok := descriptions.Endpoints[endpoint]
				if !ok {
					message = message + fmt.Sprintf("Actual description does not contain an expected endpoint. Expected: '%s'.\n", endpoint)
					continue
				}

				if !reflect.DeepEqual(expectedDesc, actualDesc) {
					message = message + fmt.Sprintf("Unexpected endpoint description. Expected: '%v', actual: '%v'.\n",
						expectedDesc, actualDesc)
				}
			}

			for currentType, expectedDesc := range expectedDescriptions.Types {
				actualDesc, ok := descriptions.Types[currentType]
				if !ok {
					message = message + fmt.Sprintf("Actual description does not contain an expected type. Expected: '%s'.\n", currentType)
					continue
				}

				if !reflect.DeepEqual(expectedDesc, actualDesc) {
					message = message + fmt.Sprintf("Unexpected type description. Expected: '%v', actual: '%v'.\n",
						expectedDesc, actualDesc)
				}
			}

			message = message + fmt.Sprintf("Unexpected API Descriptions. Expected: '%s', actual: '%s'.\n",
				util.MustToJSONIndent(expectedDescriptions), util.MustToJSONIndent(descriptions))

			test.Fatalf(message)
		}
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
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*int)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "int",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*int64)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "int64",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*bool)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "bool",
			},
			map[string]core.TypeDescription{},
		},

		// Simple wrapper types.
		{
			reflect.TypeOf((*stringWrapper)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.TypeDescription{
				getTypeID(reflect.TypeOf((*stringWrapper)(nil)).Elem()): core.TypeDescription{
					Category:  core.AliasType,
					AliasType: "string",
				},
			},
		},
		{
			reflect.TypeOf((*core.MinServerRoleAdmin)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "bool",
			},
			map[string]core.TypeDescription{
				getTypeID(reflect.TypeOf((*core.MinServerRoleAdmin)(nil)).Elem()): core.TypeDescription{
					Category:  core.AliasType,
					AliasType: "bool",
				},
			},
		},

		// Simple maps and arrays.
		{
			reflect.TypeOf((*map[string]string)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*[]string)(nil)).Elem(),
			core.TypeDescription{
				Category:    core.ArrayType,
				ElementType: "string",
			},
			map[string]core.TypeDescription{},
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
				getTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem()): core.TypeDescription{
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
				getTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem()): core.TypeDescription{
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
				Fields:   map[string]string{},
			},
			map[string]core.TypeDescription{
				getTypeID(reflect.TypeOf((*simpleStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields:   map[string]string{},
				},
			},
		},
		{
			reflect.TypeOf((*wrappedStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields:   map[string]string{},
			},
			map[string]core.TypeDescription{
				getTypeID(reflect.TypeOf((*wrappedStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields:   map[string]string{},
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
				getTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
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
				getTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
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
				getTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":      "string",
						"first-name": "string",
						"job-code":   "int",
						"last-name":  "string",
					},
				},
				getTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":    "string",
						"job-code": "int",
					},
				},
				getTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
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
					"coin-value": "github.com/edulinq/autograder/internal/api.simpleMapWrapper",
					"good-index": "github.com/edulinq/autograder/internal/api.simpleArrayWrapper",
					"personnel":  "github.com/edulinq/autograder/internal/api.embeddedJSONStruct",
				},
			},
			map[string]core.TypeDescription{
				getTypeID(reflect.TypeOf((*complexJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"coin-value": "github.com/edulinq/autograder/internal/api.simpleMapWrapper",
						"good-index": "github.com/edulinq/autograder/internal/api.simpleArrayWrapper",
						"personnel":  "github.com/edulinq/autograder/internal/api.embeddedJSONStruct",
					},
				},
				getTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":      "string",
						"first-name": "string",
						"job-code":   "int",
						"last-name":  "string",
					},
				},
				getTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"first-name": "string",
						"last-name":  "string",
					},
				},
				getTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem()): core.TypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
				getTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":    "string",
						"job-code": "int",
					},
				},
				getTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem()): core.TypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},

		// Pointers to various types.
		{
			reflect.TypeOf((**string)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((**map[string]string)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]core.TypeDescription{},
		},

		// Pointers inside of fields.
		{
			reflect.TypeOf((*simplePointerWrapper)(nil)).Elem(),
			core.TypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.TypeDescription{},
		},
		{
			reflect.TypeOf((*complexPointerStruct)(nil)).Elem(),
			core.TypeDescription{
				Category: core.StructType,
				Fields: map[string]string{
					"coin-value": "*github.com/edulinq/autograder/internal/api.simpleMapWrapper",
					"good-index": "*github.com/edulinq/autograder/internal/api.simpleArrayWrapper",
					"personnel":  "*github.com/edulinq/autograder/internal/api.embeddedJSONStruct",
				},
			},
			map[string]core.TypeDescription{
				getTypeID(reflect.TypeOf((*complexPointerStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"coin-value": "*github.com/edulinq/autograder/internal/api.simpleMapWrapper",
						"good-index": "*github.com/edulinq/autograder/internal/api.simpleArrayWrapper",
						"personnel":  "*github.com/edulinq/autograder/internal/api.embeddedJSONStruct",
					},
				},
				// Note that the keys in typeMap do not include the pointer.
				// TypeMap stores 'api/api.embeddedJSONStruct' instead of 'api/*api.embeddedJSONStruct'.
				getTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":      "string",
						"first-name": "string",
						"job-code":   "int",
						"last-name":  "string",
					},
				},
				getTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"first-name": "string",
						"last-name":  "string",
					},
				},
				getTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem()): core.TypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
				getTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem()): core.TypeDescription{
					Category: core.StructType,
					Fields: map[string]string{
						"email":    "string",
						"job-code": "int",
					},
				},
				getTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem()): core.TypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
	}

	for i, testCase := range testCases {
		typeMap := make(map[string]core.TypeDescription)

		actual, _, _ := describeType(testCase.customType, typeMap)

		if !reflect.DeepEqual(testCase.expectedDesc, actual) {
			test.Errorf("Case %d: Unexpected type simplification. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedDesc), util.MustToJSONIndent(actual))
			continue
		}

		if !reflect.DeepEqual(testCase.expectedTypeMap, typeMap) {
			test.Errorf("Case %d: Unexpected type map. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedTypeMap), util.MustToJSONIndent(typeMap))
			continue
		}
	}
}
