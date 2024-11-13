package cmd

import (
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	procedures "github.com/edulinq/autograder/internal/procedures/server"
)

var testCases = []struct {
	expectedOutputSubstring string
}{
	{API_SERVER_START},
	{UNIX_SERVER_START},
}

func TestCMDStartsServer(test *testing.T) {
	cmd := exec.Command("go", "run", "../../cmd/user-list/main.go")

	output, err := cmd.CombinedOutput()
	if err != nil {
		test.Fatal("Failed to run the CMD.", err)
	}

	for _, testCase := range testCases {
		if !strings.Contains(string(output), testCase.expectedOutputSubstring) {
			test.Error("CMD run didn't start their own server.")
		}
	}
}

func TestCMDConnectsToPrimaryServer(test *testing.T) {
	log.SetLevelFatal()

	port, err := getUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	defer config.WEB_PORT.Set(config.WEB_PORT.Get())
	config.WEB_PORT.Set(port)

	var serverStart sync.WaitGroup
	serverStart.Add(1)

	go func() {
		defer server.StopServer()

		serverStart.Done()

		err := procedures.Start(common.PrimaryServer)
		if err != nil {
			test.Fatal("Failed to start the primary server.", err)
		}
	}()

	serverStart.Wait()

	// Small sleep to let server startup
	time.Sleep(100 * time.Millisecond)

	cmd := exec.Command("go", "run", "../../cmd/user-list/main.go")

	output, err := cmd.CombinedOutput()
	if err != nil {
		test.Error("Failed to run the CMD.", err)
	}

	for _, testCase := range testCases {
		if strings.Contains(string(output), testCase.expectedOutputSubstring) {
			test.Error("CMD run started their own server when it should've connected to the primary server.")
		}
	}
}
