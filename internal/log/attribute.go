package log

import (
    "errors"
    "fmt"
)

const (
    KEY_COURSE = "course"
    KEY_ASSIGNMENT = "assignment"
    KEY_USER = "user"
)

type Attr struct {
    Name string
    Value any
}

type Loggable interface {
    LogValue() []*Attr;
}

func NewAttr(name string, value any) *Attr {
    return &Attr{name, value};
}

func NewCourseAttr(courseID string) *Attr {
    return &Attr{KEY_COURSE, courseID};
}

func NewAssignmentAttr(assignmentID string) *Attr {
    return &Attr{KEY_ASSIGNMENT, assignmentID};
}

func NewUserAttr(email string) *Attr {
    return &Attr{KEY_USER, email};
}

// Parse logging information from standard arguments.
// Arguments must be either:
// nil, error, Loggable, Attr, or *Attr.
// If there is an error parsing the attributes,
// the best effort attributes will be returned and an error will be retutned.
// Duplicate keys are ignored if the values are the same,
// and returns an error if the values are different.
func parseArgs(args ...any) (string, string, string, error, map[string]any, error) {
    var course string;
    var assignment string;
    var user string;
    var logError error;
    var attributes map[string]any = make(map[string]any);
    var err error;

    // Keep track of the index that each argement came from.
    // As we add new arguemnts from Loggables, these indexes become less clear.
    argIndexes := make([]int, len(args));
    for i, _ := range args {
        argIndexes[i] = i;
    }

    // Loop over args until there are none left.
    // We may add to them as we progress.
    for i := 0; i < len(args); i++ {
        arg := args[i];
        argIndex := argIndexes[i];

        if (arg == nil) {
            continue;
        }

        var attr *Attr = nil;

        switch argValue := arg.(type) {
            case error:
                logError = argValue;
            case Loggable:
                // Add the new values as args and continue the loop.
                for _, newArg := range argValue.LogValue() {
                    args = append(args, newArg);
                    argIndexes = append(argIndexes, i);
                }
                continue;
            case Attr:
                attr = &argValue;
            case *Attr:
                attr = argValue;
            default:
                err = errors.Join(err, fmt.Errorf("Logging argument %d is an unknown type '%T': '%v'.", argIndex, argValue, argValue));
        }

        if (attr == nil) {
            continue;
        }

        switch attr.Name {
            case KEY_COURSE:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a course key, but non-string value '%T': '%v'.", argIndex, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (value == "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is an empty course.", argIndex));
                    continue;
                }

                if (course != "") {
                    if (course != value) {
                        err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate course key. Old value: '%s', New value: '%s'.", argIndex, course, value));
                    }

                    continue;
                }

                course = value;
            case KEY_ASSIGNMENT:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a assignment key, but non-string value '%T': '%v'.", argIndex, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (value == "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is an empty assignment.", argIndex));
                    continue;
                }

                if (assignment != "") {
                    if (assignment != value) {
                        err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate assignment key. Old value: '%s', New value: '%s'.", argIndex, assignment, value));
                    }

                    continue;
                }

                assignment = value;
            case KEY_USER:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a user key, but non-string value '%T': '%v'.", argIndex, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (value == "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is an empty user.", argIndex));
                    continue;
                }

                if (user != "") {
                    if (user != value) {
                        err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate user key. Old value: '%s', New value: '%s'.", argIndex, user, value));
                    }

                    continue;
                }

                user = value;
            default:
                oldValue, exists := attributes[attr.Name];
                if (exists) {
                    if (oldValue != attr.Value) {
                        err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate key (%s) with differnet values. Old value: '%v', New value: '%v'.", argIndex, attr.Name, oldValue, attr.Value));
                    }

                    continue;
                }

                attributes[attr.Name] = attr.Value
        }
    }

    return course, assignment, user, logError, attributes, err;
}
