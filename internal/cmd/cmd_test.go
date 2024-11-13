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
	{`API Server Started.`},
	{`Unix Socket Server Started.`},
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
	// Quiet primary server startup logs.
	log.SetLevelFatal()

	port, err := getUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	defer config.WEB_PORT.Set(config.WEB_PORT.Get())
	config.WEB_PORT.Set(port)

	var serverStart sync.WaitGroup
	serverStart.Add(1)

	defer server.StopServer()
	go func() {
		serverStart.Done()

		// Mimic starting cmd/server/main.go.
		err := procedures.Start(common.PrimaryServer)
		if err != nil {
			test.Fatal("Failed to start the primary server.", err)
		}
	}()

	serverStart.Wait()

	// Small sleep to allow the server to start up.
	time.Sleep(100 * time.Millisecond)

	cmd := exec.Command("go", "run", "../../cmd/user-list/main.go")

	output, err := cmd.CombinedOutput()
	if err != nil {
		test.Error("Failed to run the CMD.", err)
	}

	for _, testCase := range testCases {
		if strings.Contains(string(output), testCase.expectedOutputSubstring) {
			test.Error("CMD run started their own server.")
		}
	}
}
