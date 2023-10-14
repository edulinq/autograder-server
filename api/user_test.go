package api

import (
    "fmt"
    "testing"

    "github.com/eriq-augustine/autograder/util"
)

// TEST - Panic test.

func TestUserAdd(test *testing.T) {
    url := serverURL + "/api/v02/user/get";

    content := map[string]any{
        "course-id": "COURSE101",
        "user-email": "admin@test.com",
        "user-pass": util.Sha256HexFromStrong("admin"),
        "email": "student@test.com",
    };

    form := map[string]string{
        API_REQUEST_CONTENT_KEY: util.MustToJSON(content),
    };

    response, err := util.Post(url, form);
    if (err != nil) {
        fmt.Println("ERROR", err);
        return;
    }

    // TEST
    fmt.Println(response);
}
