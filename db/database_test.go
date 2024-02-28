package db

import (
    "fmt"
    "reflect"
    "testing"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/log"
)

// Backends to put through the standard tests.
var testBackends []string = []string{
    DB_TYPE_DISK,
};

// Methods attatched to this struct will be called for each backend in testBackends.
type DBTests struct {
}

// Methods attatched to DBTests should follow this type.
type DBTestFunction func(*testing.T);

func TestDatabases(test *testing.T) {
    oldType := config.DB_TYPE.Get();
    defer config.DB_TYPE.Set(oldType);

    var dbTests DBTests;
    testMethods, err := getDBTests(&dbTests);
    if (err != nil) {
        test.Fatalf("Failed to get test methods: '%v'.", err);
    }

    config.MustEnableUnitTestingMode();

    // Quiet the logs.
    log.SetLevelFatal();

    for _, dbType := range testBackends {
        config.DB_TYPE.Set(dbType);

        PrepForTestingMain();

        for _, testMethod := range testMethods {
            ResetForTesting();

            test.Run(fmt.Sprintf("%s/%s", dbType, testMethod.Name), func(test *testing.T) {
                testMethod.Func.Call([]reflect.Value{reflect.ValueOf(&dbTests), reflect.ValueOf(test)});
            });
        }

        CleanupTestingMain();
    }
}

func getDBTests(dbTests *DBTests) ([]*reflect.Method, error) {
    methods := make([]*reflect.Method, 0);

    reflectValue := reflect.ValueOf(dbTests);
    reflectType := reflectValue.Type();

    for i := 0; i < reflectType.NumMethod(); i++ {
        method := reflectType.Method(i);

        // Note that the reciever counts as 1 input.
        if (method.Type.NumIn() != 2) {
            return nil, fmt.Errorf("Method '%s' does not have exactly 1 input, has %d inputs.", method.Name, method.Type.NumIn());
        }

        if (method.Type.In(0) != reflect.TypeOf((*DBTests)(nil))) {
            return nil, fmt.Errorf("Method '%s' has the wrong receiver type. Expecting '*db.DBTests', found '%s'.",
                    method.Name, method.Type.In(0));
        }

        if (method.Type.In(1) != reflect.TypeOf((*testing.T)(nil))) {
            return nil, fmt.Errorf("Method '%s' has the wrong argument type. Expecting '*testing.T', found '%s'.",
                    method.Name, method.Type.In(1));
        }

        methods = append(methods, &method);
    }

    return methods, nil;
}

// A test that does nothing.
func (this *DBTests) DBTestNoOp(test *testing.T) {
}
