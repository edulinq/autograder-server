package server

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/systemserver"
	"github.com/edulinq/autograder/internal/util"
)

const TEST_SHORT_WAIT_MS int = 50

var (
	certPath = filepath.Join(util.TestdataDirForTesting(), "certs", "ssl", "test-ssl.crt")
	keyPath  = filepath.Join(util.TestdataDirForTesting(), "certs", "ssl", "test-ssl.key")
)

func TestServerBase(test *testing.T) {
	port, err := util.GetUnusedPort()
	if err != nil {
		test.Fatalf("Failed to get an unused port: '%v'.", err)
	}

	oldPort := config.WEB_HTTP_PORT.Get()
	config.WEB_HTTP_PORT.Set(port)
	defer config.WEB_HTTP_PORT.Set(oldPort)

	// Adjust the server address for test client.
	oldURL := core.SetTestServerURL(fmt.Sprintf("http://127.0.0.1:%d", port))
	defer core.SetTestServerURL(oldURL)

	runServerTestBase(test)
}

func TestServerHTTPSBase(test *testing.T) {
	port, err := util.GetUnusedPort()
	if err != nil {
		test.Fatalf("Failed to get an unused port: '%v'.", err)
	}

	oldEnable := config.WEB_HTTPS_ENABLE.Get()
	config.WEB_HTTPS_ENABLE.Set(true)
	defer config.WEB_HTTPS_ENABLE.Set(oldEnable)

	oldCertPath := config.WEB_HTTPS_CERT.Get()
	config.WEB_HTTPS_CERT.Set(certPath)
	defer config.WEB_HTTPS_CERT.Set(oldCertPath)

	oldKeyPath := config.WEB_HTTPS_KEY.Get()
	config.WEB_HTTPS_KEY.Set(keyPath)
	defer config.WEB_HTTPS_KEY.Set(oldKeyPath)

	oldPort := config.WEB_HTTPS_PORT.Get()
	config.WEB_HTTPS_PORT.Set(port)
	defer config.WEB_HTTPS_PORT.Set(oldPort)

	// Allow insecure HTTPS requests on the client side.
	oldInsecureHTTPS := util.SetInsecureHTTPS(true)
	defer util.SetInsecureHTTPS(oldInsecureHTTPS)

	// Adjust the server address for test client.
	oldURL := core.SetTestServerURL(fmt.Sprintf("https://127.0.0.1:%d", port))
	defer core.SetTestServerURL(oldURL)

	runServerTestBase(test)
}

// Ensure that HTTP requests redirect when HTTPS (and HTTP redirect) are enabled.
func TestServerHTTPSRedirect(test *testing.T) {
	var httpPort int = 0
	var httpsPort int = 0
	var err error = nil

	// This is flakey since we may not be fast enough to get two differnt ports.
	// Try a few times before giving up.
	for i := 0; i < 10; i++ {
		httpPort, err = util.GetUnusedPort()
		if err != nil {
			test.Fatalf("Failed to get an unused http port: '%v'.", err)
		}

		httpsPort, err = util.GetUnusedPort()
		if err != nil {
			test.Fatalf("Failed to get an unused https port: '%v'.", err)
		}

		if httpPort != httpsPort {
			break
		}
	}

	if httpPort == httpsPort {
		test.Fatalf("Failed to get two different ports for HTTP and HTTPS.")
	}

	oldHTTPPort := config.WEB_HTTP_PORT.Get()
	config.WEB_HTTP_PORT.Set(httpPort)
	defer config.WEB_HTTP_PORT.Set(oldHTTPPort)

	oldHTTPSPort := config.WEB_HTTPS_PORT.Get()
	config.WEB_HTTPS_PORT.Set(httpsPort)
	defer config.WEB_HTTPS_PORT.Set(oldHTTPSPort)

	oldEnable := config.WEB_HTTPS_ENABLE.Get()
	config.WEB_HTTPS_ENABLE.Set(true)
	defer config.WEB_HTTPS_ENABLE.Set(oldEnable)

	oldRedirect := config.WEB_HTTP_REDIRECT.Get()
	config.WEB_HTTP_REDIRECT.Set(true)
	defer config.WEB_HTTP_REDIRECT.Set(oldRedirect)

	oldCertPath := config.WEB_HTTPS_CERT.Get()
	config.WEB_HTTPS_CERT.Set(certPath)
	defer config.WEB_HTTPS_CERT.Set(oldCertPath)

	oldKeyPath := config.WEB_HTTPS_KEY.Get()
	config.WEB_HTTPS_KEY.Set(keyPath)
	defer config.WEB_HTTPS_KEY.Set(oldKeyPath)

	// Allow insecure HTTPS requests on the client side.
	oldInsecureHTTPS := util.SetInsecureHTTPS(true)
	defer util.SetInsecureHTTPS(oldInsecureHTTPS)

	// Adjust the server address for test client.
	// Remember that we are making the request to HTTP so it can be redirected to HTTPS.
	oldURL := core.SetTestServerURL(fmt.Sprintf("http://127.0.0.1:%d", httpPort))
	defer core.SetTestServerURL(oldURL)

	runServerTestBase(test)
}

func runServerTestBase(test *testing.T) {
	runServerTestBaseFull(test, "courses/users/list", nil, "course-admin", "", "")
}

// All other configs should already be set.
func runServerTestBaseFull(test *testing.T, endpoint string, fields map[string]any, email string, expectedLocator string, prefix string) {
	server := NewAPIServer()
	defer server.Stop()

	var serverStopWaitGroup sync.WaitGroup
	serverStopWaitGroup.Add(1)

	// Run the server in the background.
	go func() {
		defer serverStopWaitGroup.Done()

		err := server.RunAndBlock(systemserver.CMD_TEST_SERVER)
		if err != nil {
			test.Fatalf("The server returned an error: '%v'.", err)
		}
	}()

	// Wait for the server to start.
	time.Sleep(time.Duration(TEST_SHORT_WAIT_MS) * time.Millisecond)

	// Send a request.
	response := core.SendTestAPIRequestFull(test, endpoint, fields, nil, email)
	if !response.Success {
		if expectedLocator != "" {
			if expectedLocator != response.Locator {
				test.Fatalf("%sIncorrect locator. Expected: '%s', Actual: '%s'.", prefix, expectedLocator, response.Locator)
			}
		} else {
			test.Fatalf("%sResponse is not a success when it should be: '%v'.", prefix, response)
		}
	}

	// Wait for the server to stop.
	server.Stop()
	serverStopWaitGroup.Wait()

	// Wait for a small time to ensure all stopping operations are done.
	time.Sleep(time.Duration(TEST_SHORT_WAIT_MS) * time.Millisecond)
}
