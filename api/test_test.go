package api

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/util"
)

var server *httptest.Server;
var serverURL string;

func startTestServer() {
    if (server != nil) {
        panic("Test server already started.");
    }

    server = httptest.NewServer(http.HandlerFunc(serve));
    serverURL = server.URL;
}

func stopTestServer() {
    if (server != nil) {
        server.Close();

        server = nil;
        serverURL = "";
    }
}

// Make a request to the test server using fields for
// a standard test request plus whatever other fields are specified.
// Provided fields will override base fields.
func sendTestAPIRequest(test *testing.T, endpoint string, fields map[string]any) *APIResponse {
    url := serverURL + endpoint;

    content := map[string]any{
        "course-id": "COURSE101",
        "user-email": "admin@test.com",
        "user-pass": util.Sha256HexFromStrong("admin"),
    };

    for key, value := range fields {
        content[key] = value;
    }

    form := map[string]string{
        API_REQUEST_CONTENT_KEY: util.MustToJSON(content),
    };

    responseText, err := util.PostNoCheck(url, form);
    if (err != nil) {
        test.Fatalf("API POST returned an error: '%v'.", err);
    }

    var response APIResponse;
    err = util.JSONFromString(responseText, &response);
    if (err != nil) {
        test.Fatalf("Could not unmarshal JSON response '%s': '%v'.", responseText, err);
    }

    return &response;
}

// Common setup for all API tests.
func TestMain(suite *testing.M) {
    config.EnableTestingMode(false, true);
    config.NO_AUTH.Set(false);

    err := grader.LoadCourses();
    if (err != nil) {
        fmt.Printf("Failed to load test courses: '%v'.", err);
        os.Exit(1);
    }

    startTestServer();
    defer stopTestServer();

    os.Exit(suite.Run())
}
