package api

import (
    "fmt"
    "testing"
)

// TEST
func TestUserAdd(test *testing.T) {
    fields := map[string]any{
        "email": "student@test.com",
    };

    response := sendTestAPIRequest(test, "/api/v02/user/get", fields);

    // TEST
    fmt.Println(response);
}
