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
    LogValue() *Attr;
}

func NewAttr(name string, value any) *Attr {
    return &Attr{name, value};
}

// Parse logging information from standard arguments.
// Arguments must be either:
// nil, error, Loggable, Attr, or *Attr.
// If there is an error parsing the attributes,
// the best effort attributes will be returned and an error will be retutned.
func parseArgs(args ...any) (string, string, string, error, map[string]any, error) {
    var course string;
    var assignment string;
    var user string;
    var logError error;
    var attributes map[string]any = make(map[string]any);
    var err error;

    for i, arg := range args {
        if (arg == nil) {
            continue;
        }

        var attr *Attr = nil;

        switch argValue := arg.(type) {
            case error:
                logError = argValue;
            case Loggable:
                attr = argValue.LogValue();
            case Attr:
                attr = &argValue;
            case *Attr:
                attr = argValue;
            default:
                err = errors.Join(err, fmt.Errorf("Logging argument %d is an unknown type '%T': '%v'.", i, argValue, argValue));
        }

        if (attr == nil) {
            continue;
        }

        switch attr.Name {
            case KEY_COURSE:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a course key, but non-string value '%T': '%v'.", i, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (course != "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate course key. Old value: '%s', New value: '%s'.", i, course, value));
                }

                course = value;
            case KEY_ASSIGNMENT:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a assignment key, but non-string value '%T': '%v'.", i, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (assignment != "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate assignment key. Old value: '%s', New value: '%s'.", i, assignment, value));
                }

                assignment = value;
            case KEY_USER:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a user key, but non-string value '%T': '%v'.", i, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (user != "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate user key. Old value: '%s', New value: '%s'.", i, user, value));
                }

                user = value;
            default:
                attributes[attr.Name] = attr.Value
        }
    }

    return course, assignment, user, logError, attributes, err;
}
