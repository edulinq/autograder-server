package api

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
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
