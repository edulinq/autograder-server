package util

import (
    "fmt"
    "reflect"
)

func BaseString(obj any) string {
    json, err := ToJSON(obj);
    if (err != nil) {
        // Explicitly use Go-Syntax (%#v) to avoid loops with overwritten String() methods.
        return fmt.Sprintf("%#v", reflect.ValueOf(obj));
    }

    return json;
}
