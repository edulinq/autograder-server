package core

import (
	"reflect"
	"testing"

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

func TestDescribeRoutesEmptyRoutes(test *testing.T) {
	routes := []Route{}
	description, err := DescribeRoutes(routes)
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	if len(description.Endpoints) != 0 {
		test.Errorf("Unexpected number of endpoints. Expected: '0', actual: '%d'.", len(description.Endpoints))
	}
}

func TestDescribeTypeBase(test *testing.T) {
	testCases := []struct {
		customType      reflect.Type
		expectedDesc    TypeDescription
		expectedTypeMap map[string]TypeDescription
	}{
		// Base types to alias (no JSON tags).
		{
			reflect.TypeOf((*string)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "string",
			},
			map[string]TypeDescription{},
		},
		{
			reflect.TypeOf((*int)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "int",
			},
			map[string]TypeDescription{},
		},
		{
			reflect.TypeOf((*int64)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "int64",
			},
			map[string]TypeDescription{},
		},
		{
			reflect.TypeOf((*bool)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "bool",
			},
			map[string]TypeDescription{},
		},

		// Simple wrapper types.
		{
			reflect.TypeOf((*stringWrapper)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "string",
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*stringWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:  AliasType,
					AliasType: "string",
				},
			},
		},
		{
			reflect.TypeOf((*MinServerRoleAdmin)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "bool",
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*MinServerRoleAdmin)(nil)).Elem(), nil): TypeDescription{
					Category:  AliasType,
					AliasType: "bool",
				},
			},
		},

		// Simple maps and arrays.
		{
			reflect.TypeOf((*map[string]string)(nil)).Elem(),
			TypeDescription{
				Category:  MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]TypeDescription{},
		},
		{
			reflect.TypeOf((*[]string)(nil)).Elem(),
			TypeDescription{
				Category:    ArrayType,
				ElementType: "string",
			},
			map[string]TypeDescription{},
		},

		// Wrapped maps and arrays.
		{
			reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(),
			TypeDescription{
				Category:  MapType,
				KeyType:   "string",
				ValueType: "int",
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:  MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
		{
			reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(),
			TypeDescription{
				Category:    ArrayType,
				ElementType: "bool",
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:    ArrayType,
					ElementType: "bool",
				},
			},
		},

		// Fields without JSON tags are ignored.
		{
			reflect.TypeOf((*simpleStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields:   []FieldDescription{},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields:   []FieldDescription{},
				},
			},
		},
		{
			reflect.TypeOf((*wrappedStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields:   []FieldDescription{},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*wrappedStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields:   []FieldDescription{},
				},
			},
		},

		// Simple JSON tags.
		{
			reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields: []FieldDescription{
					FieldDescription{"email", "string"},
					FieldDescription{"job-code", "int"},
				},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"job-code", "int"},
					},
				},
			},
		},

		// Hidden JSON tags (-).
		{
			reflect.TypeOf((*secureJSONStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields: []FieldDescription{
					FieldDescription{"first-name", "string"},
					FieldDescription{"last-name", "string"},
				},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"first-name", "string"},
						FieldDescription{"last-name", "string"},
					},
				},
			},
		},

		// Embedded fields.
		{
			reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields: []FieldDescription{
					FieldDescription{"email", "string"},
					FieldDescription{"first-name", "string"},
					FieldDescription{"job-code", "int"},
					FieldDescription{"last-name", "string"},
				},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"first-name", "string"},
						FieldDescription{"job-code", "int"},
						FieldDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"job-code", "int"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"first-name", "string"},
						FieldDescription{"last-name", "string"},
					},
				},
			},
		},

		// Complex fields.
		{
			reflect.TypeOf((*complexJSONStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields: []FieldDescription{
					FieldDescription{"coin-value", "core.simpleMapWrapper"},
					FieldDescription{"good-index", "core.simpleArrayWrapper"},
					FieldDescription{"personnel", "core.embeddedJSONStruct"},
				},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*complexJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"coin-value", "core.simpleMapWrapper"},
						FieldDescription{"good-index", "core.simpleArrayWrapper"},
						FieldDescription{"personnel", "core.embeddedJSONStruct"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"first-name", "string"},
						FieldDescription{"job-code", "int"},
						FieldDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"first-name", "string"},
						FieldDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:    ArrayType,
					ElementType: "bool",
				},
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"job-code", "int"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:  MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},

		// Pointers to various types.
		{
			reflect.TypeOf((**string)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "string",
			},
			map[string]TypeDescription{},
		},
		{
			reflect.TypeOf((**map[string]string)(nil)).Elem(),
			TypeDescription{
				Category:  MapType,
				KeyType:   "string",
				ValueType: "string",
			},
			map[string]TypeDescription{},
		},

		// Pointers inside of fields.
		{
			reflect.TypeOf((*simplePointerWrapper)(nil)).Elem(),
			TypeDescription{
				Category:  AliasType,
				AliasType: "string",
			},
			map[string]TypeDescription{},
		},
		{
			reflect.TypeOf((*complexPointerStruct)(nil)).Elem(),
			TypeDescription{
				Category: StructType,
				Fields: []FieldDescription{
					FieldDescription{"coin-value", "*core.simpleMapWrapper"},
					FieldDescription{"good-index", "*core.simpleArrayWrapper"},
					FieldDescription{"personnel", "*core.embeddedJSONStruct"},
				},
			},
			map[string]TypeDescription{
				mustGetTypeID(reflect.TypeOf((*complexPointerStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"coin-value", "*core.simpleMapWrapper"},
						FieldDescription{"good-index", "*core.simpleArrayWrapper"},
						FieldDescription{"personnel", "*core.embeddedJSONStruct"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"first-name", "string"},
						FieldDescription{"job-code", "int"},
						FieldDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"first-name", "string"},
						FieldDescription{"last-name", "string"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:    ArrayType,
					ElementType: "bool",
				},
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): TypeDescription{
					Category: StructType,
					Fields: []FieldDescription{
						FieldDescription{"email", "string"},
						FieldDescription{"job-code", "int"},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): TypeDescription{
					Category:  MapType,
					KeyType:   "string",
					ValueType: "int",
				},
			},
		},
	}

	for i, testCase := range testCases {
		info := TypeInfoCache{
			TypeMap: make(map[string]TypeDescription),
		}

		actual, _, _, err := DescribeType(testCase.customType, true, info)
		if err != nil {
			test.Errorf("Case %d: Unexpected error while describing types: '%v'.", i, err)
		}

		if !reflect.DeepEqual(testCase.expectedDesc, actual) {
			test.Errorf("Case %d: Unexpected type simplification. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedDesc), util.MustToJSONIndent(actual))
			continue
		}

		if !reflect.DeepEqual(testCase.expectedTypeMap, info.TypeMap) {
			test.Errorf("Case %d: Unexpected type map. Expected: '%v', actual: '%v'.",
				i, util.MustToJSONIndent(testCase.expectedTypeMap), util.MustToJSONIndent(info.TypeMap))
			continue
		}
	}
}
