package core

import (
    "net/http/httptest"
    "os"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

var server *httptest.Server;
var serverURL string;

func startTestServer(routes *[]*Route) {
    if (server != nil) {
        panic("Test server already started.");
    }

    server = httptest.NewServer(GetRouteServer(routes));
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
func APITestingMain(suite *testing.M, routes *[]*Route) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        db.PrepForTestingMain();
        defer db.CleanupTestingMain();

        config.NO_AUTH.Set(false);

        startTestServer(routes);
        defer stopTestServer();

        return suite.Run();
    }();

    os.Exit(code);
}

func SendTestAPIRequest(test *testing.T, endpoint string, fields map[string]any) *APIResponse {
    return SendTestAPIRequestFull(test, endpoint, fields, nil, usr.Admin);
}

// Make a request to the test server using fields for
// a standard test request plus whatever other fields are specified.
// Provided fields will override base fields.
// The given role will choose the user (the test course has one user per role).
func SendTestAPIRequestFull(test *testing.T, endpoint string, fields map[string]any, paths []string, role usr.UserRole) *APIResponse {
    url := serverURL + endpoint;

    email := usr.GetRoleString(role) + "@test.com";
    pass := util.Sha256HexFromString(usr.GetRoleString(role));

    content := map[string]any{
        "course-id": "course101",
        "assignment-id": "hw0",
        "user-email": email,
        "user-pass": pass,
    };

    for key, value := range fields {
        content[key] = value;
    }

    form := map[string]string{
        API_REQUEST_CONTENT_KEY: util.MustToJSON(content),
    };

    var responseText string;
    var err error;

    if (len(paths) == 0) {
        responseText, err = util.PostNoCheck(url, form);
    } else {
        responseText, err = util.PostFiles(url, form, paths, false);
    }

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
