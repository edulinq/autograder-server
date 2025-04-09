package api

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	courseUsers "github.com/edulinq/autograder/internal/api/courses/users"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/log"
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

func mustGetTypeID(customType reflect.Type, typeConversions map[string]string) string {
	typeID, err := getTypeID(customType, typeConversions)
	if err != nil {
		log.Fatal("Failed to get type ID.", err)
	}

	return typeID
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
			message := "Unexpected API Descriptions.\n"

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

			test.Fatal(message)
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

func TestDescribeType(test *testing.T) {
	testCases := []struct {
		customType      reflect.Type
		expectedDesc    core.FullTypeDescription
		expectedTypeMap map[string]core.FullTypeDescription
	}{
		// Base types to alias (no JSON tags).
		{
			reflect.TypeOf((*string)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.FullTypeDescription{},
		},
		{
			reflect.TypeOf((*int)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "int",
			},
			map[string]core.FullTypeDescription{},
		},
		{
			reflect.TypeOf((*int64)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "int64",
			},
			map[string]core.FullTypeDescription{},
		},
		{
			reflect.TypeOf((*bool)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "bool",
			},
			map[string]core.FullTypeDescription{},
		},

		// Simple wrapper types.
		{
			reflect.TypeOf((*stringWrapper)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*stringWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:  core.AliasType,
					AliasType: "string",
				},
			},
		},
		{
			reflect.TypeOf((*core.MinServerRoleAdmin)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "bool",
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*core.MinServerRoleAdmin)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:  core.AliasType,
					AliasType: "bool",
				},
			},
		},

		// Simple maps and arrays.
		{
			reflect.TypeOf((*map[string]string)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]core.FullTypeDescription{},
		},
		{
			reflect.TypeOf((*[]string)(nil)).Elem(),
			core.FullTypeDescription{
				Category:    core.ArrayType,
				ElementType: "string",
			},
			map[string]core.FullTypeDescription{},
		},

		// Wrapped maps and arrays.
		{
			reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "int",
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
		{
			reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(),
			core.FullTypeDescription{
				Category:    core.ArrayType,
				ElementType: "bool",
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
			},
		},

		// Fields without JSON tags are ignored.
		{
			reflect.TypeOf((*simpleStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields:   []core.TypeDescription{},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields:   []core.TypeDescription{},
				},
			},
		},
		{
			reflect.TypeOf((*wrappedStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields:   []core.TypeDescription{},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*wrappedStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields:   []core.TypeDescription{},
				},
			},
		},

		// Simple JSON tags.
		{
			reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields: []core.TypeDescription{
					core.TypeDescription{"email", "string"},
					core.TypeDescription{"job-code", "int"},
				},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"job-code", "int"},
					},
				},
			},
		},

		// Hidden JSON tags (-).
		{
			reflect.TypeOf((*secureJSONStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields: []core.TypeDescription{
					core.TypeDescription{"first-name", "string"},
					core.TypeDescription{"last-name", "string"},
				},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"last-name", "string"},
					},
				},
			},
		},

		// Embedded fields.
		{
			reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields: []core.TypeDescription{
					core.TypeDescription{"email", "string"},
					core.TypeDescription{"first-name", "string"},
					core.TypeDescription{"job-code", "int"},
					core.TypeDescription{"last-name", "string"},
				},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"job-code", "int"},
						core.TypeDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"job-code", "int"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"last-name", "string"},
					},
				},
			},
		},

		// Complex fields.
		{
			reflect.TypeOf((*complexJSONStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields: []core.TypeDescription{
					core.TypeDescription{"coin-value", "api.simpleMapWrapper"},
					core.TypeDescription{"good-index", "api.simpleArrayWrapper"},
					core.TypeDescription{"personnel", "api.embeddedJSONStruct"},
				},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*complexJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"coin-value", "api.simpleMapWrapper"},
						core.TypeDescription{"good-index", "api.simpleArrayWrapper"},
						core.TypeDescription{"personnel", "api.embeddedJSONStruct"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"job-code", "int"},
						core.TypeDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"job-code", "int"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},

		// Pointers to various types.
		{
			reflect.TypeOf((**string)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.FullTypeDescription{},
		},
		{
			reflect.TypeOf((**map[string]string)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]core.FullTypeDescription{},
		},

		// Pointers inside of fields.
		{
			reflect.TypeOf((*simplePointerWrapper)(nil)).Elem(),
			core.FullTypeDescription{
				Category:  core.AliasType,
				AliasType: "string",
			},
			map[string]core.FullTypeDescription{},
		},
		{
			reflect.TypeOf((*complexPointerStruct)(nil)).Elem(),
			core.FullTypeDescription{
				Category: core.StructType,
				Fields: []core.TypeDescription{
					core.TypeDescription{"coin-value", "*api.simpleMapWrapper"},
					core.TypeDescription{"good-index", "*api.simpleArrayWrapper"},
					core.TypeDescription{"personnel", "*api.embeddedJSONStruct"},
				},
			},
			map[string]core.FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*complexPointerStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"coin-value", "*api.simpleMapWrapper"},
						core.TypeDescription{"good-index", "*api.simpleArrayWrapper"},
						core.TypeDescription{"personnel", "*api.embeddedJSONStruct"},
					},
				},
				// Note that the keys in typeMap do not include the pointer.
				// TypeMap stores 'api/api.embeddedJSONStruct' instead of 'api/*api.embeddedJSONStruct'.
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"job-code", "int"},
						core.TypeDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"first-name", "string"},
						core.TypeDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:    core.ArrayType,
					ElementType: "bool",
				},
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): core.FullTypeDescription{
					Category: core.StructType,
					Fields: []core.TypeDescription{
						core.TypeDescription{"email", "string"},
						core.TypeDescription{"job-code", "int"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): core.FullTypeDescription{
					Category:  core.MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
	}

	for i, testCase := range testCases {
		typeMap := make(map[string]core.FullTypeDescription)

		actual, _, _, _, err := describeType(testCase.customType, true, typeMap, nil)
		if err != nil {
			test.Errorf("Case %d: Unexpected error while describing types: '%v'.", i, err)
		}

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

func TestDescribeConflictingTypes(test *testing.T) {
	typeMap := make(map[string]core.FullTypeDescription)
	typeDescriptions := make(map[string]string)

	// Add in the first users.ListRequest which will work.
	_, _, _, _, err := describeType(reflect.TypeOf((*users.ListRequest)(nil)).Elem(), true, typeMap, typeDescriptions)
	if err != nil {
		test.Fatalf("Failed to describe type: '%v'.", err)
	}

	// Add in the second users.ListRequest which will cause a conflict.
	_, _, _, _, err = describeType(reflect.TypeOf((*courseUsers.ListRequest)(nil)).Elem(), true, typeMap, typeDescriptions)
	if err == nil {
		test.Fatalf("Did not get expected error while describing types.")
	}

	expectedMessage := "Unable to get type ID due to conflicting names."
	if !strings.Contains(err.Error(), expectedMessage) {
		test.Fatalf("Did not get the expected error output. Expected substring: '%s', actual: '%s'.",
			expectedMessage, err.Error())
	}
}
