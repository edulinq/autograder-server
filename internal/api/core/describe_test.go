package core

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// This is a simple alias for the string type.
type stringWrapper string

// A named map type.
type simpleMapWrapper map[string]int

// Wrapping a list in a named type!
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
	Email string `json:"email"`

	// The job code for the employee.
	JobCode int `json:"job-code"`
}

type secureJSONStruct struct {
	FirstName string `json:"first-name" required:""`
	LastName  string `json:"last-name" required:""`
	Pay       int    `json:"-"`
}

type embeddedJSONStruct struct {
	simpleJSONStruct
	secureJSONStruct
}

type complexJSONStruct struct {
	// The value of the coin.
	CoinValue simpleMapWrapper   `json:"coin-value"`
	GoodIndex simpleArrayWrapper `json:"good-index"`
	Personnel embeddedJSONStruct `json:"personnel"`
}

type complexPointerStruct struct {
	// The value of the coin.
	// A nil value indicates an unknown value.
	CoinValue *simpleMapWrapper   `json:"coin-value"`
	GoodIndex *simpleArrayWrapper `json:"good-index"`
	Personnel *embeddedJSONStruct `json:"personnel"`
}

type typeOverrideStruct struct {
	// Note: Only the overriden type name appears.
	MinServerRoleAdmin `json:"type-override-struct"`
}

type fullTypeOverrideStruct struct {
	// Note: The embedded struct name still appears as the field name.
	MinServerRoleAdmin
}

type ignoredTypeOverride struct {
	MinServerRoleAdmin `json:"-"`
}

type errorTypeFalse struct {
	badTag string `json:"bad-tag" required:"false"`
}

type errorTypeTrue struct {
	badTag string `json:"bad-tag" required:"true"`
}

//	__TYPE_DESCRIPTION_OVERRIDE__ {
//	    "category": "override",
//	    "description": "This type is missing an override type ID."
//	}
type errorTypeOverrideID bool

// __TYPE_DESCRIPTION_OVERRIDE__ "error-type-override-type"
type errorTypeOverrideType bool

//	__TYPE_DESCRIPTION_OVERRIDE__ "error-type-override-JSON" = {
//	    "category": "override",
//	    "description": "This type has an invalid JSON structure.",
//	}
type errorTypeOverrideJSON bool

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
		expectedDesc    FullTypeDescription
		expectedTypeMap map[string]FullTypeDescription
		errorSubstring  string
	}{
		// Base types to alias (no JSON tags).
		{
			reflect.TypeOf((*string)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  AliasType,
					AliasType: "string",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},
		{
			reflect.TypeOf((*int)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  AliasType,
					AliasType: "int",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},
		{
			reflect.TypeOf((*int64)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  AliasType,
					AliasType: "int64",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},
		{
			reflect.TypeOf((*bool)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  AliasType,
					AliasType: "bool",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},

		// Simple wrapper types.
		{
			reflect.TypeOf((*stringWrapper)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Description: "This is a simple alias for the string type.",
					Category:    AliasType,
					AliasType:   "string",
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*stringWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "This is a simple alias for the string type.",
						Category:    AliasType,
						AliasType:   "string",
					},
				},
			},
			"",
		},

		// Simple maps and arrays.
		{
			reflect.TypeOf((*map[string]string)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  MapType,
					KeyType:   "string",
					ValueType: "string",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},
		{
			reflect.TypeOf((*[]string)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:    ArrayType,
					ElementType: "string",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},

		// Wrapped maps and arrays.
		{
			reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Description: "A named map type.",
					Category:    MapType,
					KeyType:     "string",
					ValueType:   "int",
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "A named map type.",
						Category:    MapType,
						KeyType:     "string",
						ValueType:   "int",
					},
				},
			},
			"",
		},
		{
			reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Description: "Wrapping a list in a named type!",
					Category:    ArrayType,
					ElementType: "bool",
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "Wrapping a list in a named type!",
						Category:    ArrayType,
						ElementType: "bool",
					},
				},
			},
			"",
		},

		// Fields without JSON tags are ignored.
		{
			reflect.TypeOf((*simpleStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{},
				},
			},
			"",
		},
		{
			reflect.TypeOf((*wrappedStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*wrappedStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{},
				},
			},
			"",
		},

		// Simple JSON tags.
		{
			reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "email",
							Type: "string",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name:        "job-code",
							Type:        "int",
							Description: "The job code for the employee.",
						},
						Required: false,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*simpleJSONStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "email",
								Type: "string",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "job-code",
								Type:        "int",
								Description: "The job code for the employee.",
							},
							Required: false,
						},
					},
				},
			},
			"",
		},

		// Hidden JSON tags (-).
		{
			reflect.TypeOf((*secureJSONStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "first-name",
							Type: "string",
						},
						Required: true,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "last-name",
							Type: "string",
						},
						Required: true,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*secureJSONStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "first-name",
								Type: "string",
							},
							Required: true,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "last-name",
								Type: "string",
							},
							Required: true,
						},
					},
				},
			},
			"",
		},

		// Embedded fields.
		{
			reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "email",
							Type: "string",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "first-name",
							Type: "string",
						},
						Required: true,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name:        "job-code",
							Type:        "int",
							Description: "The job code for the employee.",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "last-name",
							Type: "string",
						},
						Required: true,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "email",
								Type: "string",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "first-name",
								Type: "string",
							},
							Required: true,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "job-code",
								Type:        "int",
								Description: "The job code for the employee.",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "last-name",
								Type: "string",
							},
							Required: true,
						},
					},
				},
			},
			"",
		},

		// Complex fields.
		{
			reflect.TypeOf((*complexJSONStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name:        "coin-value",
							Type:        "core.simpleMapWrapper",
							Description: "The value of the coin.",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "good-index",
							Type: "core.simpleArrayWrapper",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "personnel",
							Type: "core.embeddedJSONStruct",
						},
						Required: false,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*complexJSONStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "coin-value",
								Type:        "core.simpleMapWrapper",
								Description: "The value of the coin.",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "good-index",
								Type: "core.simpleArrayWrapper",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "personnel",
								Type: "core.embeddedJSONStruct",
							},
							Required: false,
						},
					},
				},
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "email",
								Type: "string",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "first-name",
								Type: "string",
							},
							Required: true,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "job-code",
								Type:        "int",
								Description: "The job code for the employee.",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "last-name",
								Type: "string",
							},
							Required: true,
						},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "Wrapping a list in a named type!",
						Category:    ArrayType,
						ElementType: "bool",
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "A named map type.",
						Category:    MapType,
						KeyType:     "string",
						ValueType:   "int",
					},
				},
			},
			"",
		},

		// Pointers to various types.
		{
			reflect.TypeOf((**string)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  AliasType,
					AliasType: "string",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},
		{
			reflect.TypeOf((**map[string]string)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  MapType,
					KeyType:   "string",
					ValueType: "string",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},

		// Pointers inside of fields.
		{
			reflect.TypeOf((*simplePointerWrapper)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:  AliasType,
					AliasType: "string",
				},
			},
			map[string]FullTypeDescription{},
			"",
		},
		{
			reflect.TypeOf((*complexPointerStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: StructType,
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name:        "coin-value",
							Type:        "*core.simpleMapWrapper",
							Description: "The value of the coin.\nA nil value indicates an unknown value.",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "good-index",
							Type: "*core.simpleArrayWrapper",
						},
						Required: false,
					},
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name: "personnel",
							Type: "*core.embeddedJSONStruct",
						},
						Required: false,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*complexPointerStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "coin-value",
								Type:        "*core.simpleMapWrapper",
								Description: "The value of the coin.\nA nil value indicates an unknown value.",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "good-index",
								Type: "*core.simpleArrayWrapper",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "personnel",
								Type: "*core.embeddedJSONStruct",
							},
							Required: false,
						},
					},
				},
				mustGetTypeID(reflect.TypeOf((*embeddedJSONStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: StructType,
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "email",
								Type: "string",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "first-name",
								Type: "string",
							},
							Required: true,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "job-code",
								Type:        "int",
								Description: "The job code for the employee.",
							},
							Required: false,
						},
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name: "last-name",
								Type: "string",
							},
							Required: true,
						},
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleArrayWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "Wrapping a list in a named type!",
						Category:    ArrayType,
						ElementType: "bool",
					},
				},
				mustGetTypeID(reflect.TypeOf((*simpleMapWrapper)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Description: "A named map type.",
						Category:    MapType,
						KeyType:     "string",
						ValueType:   "int",
					},
				},
			},
			"",
		},

		// Type overrides.
		{
			reflect.TypeOf((*MinServerRoleAdmin)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category:    "role",
					Description: "The requesting user must have a minimum server role of admin to complete this operation.",
				},
			},
			map[string]FullTypeDescription{
				"min-server-role-admin": FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category:    "role",
						Description: "The requesting user must have a minimum server role of admin to complete this operation.",
					},
				},
			},
			"",
		},
		{
			reflect.TypeOf((*typeOverrideStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: "struct",
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name:        "type-override-struct",
							Type:        "min-server-role-admin",
							Description: "Note: Only the overriden type name appears.",
						},
						Required: false,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*typeOverrideStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: "struct",
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "type-override-struct",
								Type:        "min-server-role-admin",
								Description: "Note: Only the overriden type name appears.",
							},
							Required: false,
						},
					},
				},
				"min-server-role-admin": FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category:    "role",
						Description: "The requesting user must have a minimum server role of admin to complete this operation.",
					},
				},
			},
			"",
		},
		{
			reflect.TypeOf((*fullTypeOverrideStruct)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: "struct",
				},
				Fields: []FieldDescription{
					FieldDescription{
						BaseFieldDescription: BaseFieldDescription{
							Name:        "MinServerRoleAdmin",
							Type:        "min-server-role-admin",
							Description: "Note: The embedded struct name still appears as the field name.",
						},
						Required: false,
					},
				},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*fullTypeOverrideStruct)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: "struct",
					},
					Fields: []FieldDescription{
						FieldDescription{
							BaseFieldDescription: BaseFieldDescription{
								Name:        "MinServerRoleAdmin",
								Type:        "min-server-role-admin",
								Description: "Note: The embedded struct name still appears as the field name.",
							},
							Required: false,
						},
					},
				},
				"min-server-role-admin": FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category:    "role",
						Description: "The requesting user must have a minimum server role of admin to complete this operation.",
					},
				},
			},
			"",
		},

		// Ignored type overrides.
		{
			reflect.TypeOf((*ignoredTypeOverride)(nil)).Elem(),
			FullTypeDescription{
				BaseTypeDescription: BaseTypeDescription{
					Category: "struct",
				},
				Fields: []FieldDescription{},
			},
			map[string]FullTypeDescription{
				mustGetTypeID(reflect.TypeOf((*ignoredTypeOverride)(nil)).Elem(), nil): FullTypeDescription{
					BaseTypeDescription: BaseTypeDescription{
						Category: "struct",
					},
					Fields: []FieldDescription{},
				},
			},
			"",
		},

		// Errors.
		{
			reflect.TypeOf((*errorTypeFalse)(nil)).Elem(),
			FullTypeDescription{},
			nil,
			"Unexpected required tag value. Expected: '', Actual: 'false'.",
		},
		{
			reflect.TypeOf((*errorTypeTrue)(nil)).Elem(),
			FullTypeDescription{},
			nil,
			"Unexpected required tag value. Expected: '', Actual: 'true'.",
		},
		{
			reflect.TypeOf((*errorTypeOverrideID)(nil)).Elem(),
			FullTypeDescription{},
			nil,
			"Unexpected description override. Expected: '<typeID>=<typeDescription>', Actual:",
		},
		{
			reflect.TypeOf((*errorTypeOverrideType)(nil)).Elem(),
			FullTypeDescription{},
			nil,
			"Unexpected description override. Expected: '<typeID>=<typeDescription>', Actual:",
		},
		{
			reflect.TypeOf((*errorTypeOverrideJSON)(nil)).Elem(),
			FullTypeDescription{},
			nil,
			"Failed to convert type override: 'Could not unmarshal JSON bytes/string",
		},
	}

	for i, testCase := range testCases {
		info := TypeInfoCache{
			TypeMap: make(map[string]FullTypeDescription),
		}

		actual, _, _, err := DescribeType(testCase.customType, true, info)
		if err != nil {
			if testCase.errorSubstring == "" {
				test.Errorf("Case %d: Unexpected error while describing types: '%v'.", i, err)
				continue
			}

			if !strings.Contains(err.Error(), testCase.errorSubstring) {
				test.Errorf("Case %d: Error is not as expected. Expected Substring: '%s', Actual: '%s'.",
					i, testCase.errorSubstring, err.Error())
				continue
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error '%s' on '%+v'.", i, testCase.errorSubstring, testCase.customType)
			continue
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
