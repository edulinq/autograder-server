package canvas

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

const (
	TEST_COURSE_ID     = "12345"
	TEST_ASSIGNMENT_ID = "98765"
	TEST_TOKEN         = "ABC123"
)

var server *httptest.Server
var serverURL string

//go:embed testdata/http
var httpDataDir embed.FS

var testBackend *CanvasBackend

func TestMain(suite *testing.M) {
	var err error

	// Run inside a func so defers will run before os.Exit().
	code := func() int {
		db.PrepForTestingMain()
		defer db.CleanupTestingMain()

		err = startTestServer()
		if err != nil {
			panic(err)
		}
		defer stopTestServer()

		testBackend, err = NewBackend(TEST_COURSE_ID, TEST_TOKEN, serverURL)
		if err != nil {
			panic(err)
		}

		return suite.Run()
	}()

	os.Exit(code)
}

func startTestServer() error {
	if server != nil {
		return fmt.Errorf("Test server already started.")
	}

	requests, err := loadRequests()
	if err != nil {
		return err
	}

	server = httptest.NewServer(makeHandler(requests))
	serverURL = server.URL

	return nil
}

func makeHandler(requests map[string]*common.SavedHTTPRequest) http.Handler {
	return &testCanvasHandler{requests}
}

type testCanvasHandler struct {
	requests map[string]*common.SavedHTTPRequest
}

func (this *testCanvasHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	key := fmt.Sprintf("%s::%s?%s", request.Method, request.URL.Path, request.URL.RawQuery)
	savedRequest := this.requests[key]
	if savedRequest == nil {
		fmt.Printf("ERROR 404: '%s'.\n", key)
		http.NotFound(response, request)
		return
	}

	for key, value := range savedRequest.ResponseHeaders {
		response.Header()[key] = value
	}

	response.WriteHeader(savedRequest.ResponseCode)
	_, err := response.Write([]byte(savedRequest.ResponseBody))
	if err != nil {
		panic(err)
	}
}

func loadRequests() (map[string]*common.SavedHTTPRequest, error) {
	requests := make(map[string]*common.SavedHTTPRequest)

	err := fs.WalkDir(httpDataDir, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		data, err := httpDataDir.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Failed to read embedded test file '%s': '%w'.", path, err)
		}

		var request common.SavedHTTPRequest
		err = util.JSONFromString(string(data), &request)
		if err != nil {
			return fmt.Errorf("Failed to JSON parse test file '%s': '%w'.", path, err)
		}

		uri, err := url.Parse(request.URL)
		if err != nil {
			return fmt.Errorf("Failed to parse test URL '%s': '%w'.", request.URL, err)
		}

		key := fmt.Sprintf("%s::%s?%s", request.Method, uri.Path, uri.RawQuery)
		requests[key] = &request

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to walk embeded test dir: '%w'.", err)
	}

	return requests, nil
}

func stopTestServer() {
	if server != nil {
		server.Close()

		server = nil
		serverURL = ""
	}
}
