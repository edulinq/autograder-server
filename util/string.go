package util

import (
    "fmt"
)

func BaseString(obj any) string {
    json, err := ToJSON(obj);
    if (err != nil) {
        return fmt.Sprintf("%v", obj);
    }

    return json;
}
