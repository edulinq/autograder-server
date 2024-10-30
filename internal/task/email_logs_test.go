package task

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model/tasks"
)

func TestEmailLogsBase(test *testing.T) {
	to := []string{"test1@test.edulinq.org", "test2@test.edulinq.org"}
	query := log.RawLogQuery{}

	runTest(test, to, false, query, 1, 3)
}

func TestEmailLogsEmptyNoEmail(test *testing.T) {
	to := []string{"test1@test.edulinq.org", "test2@test.edulinq.org"}
	query := log.RawLogQuery{
		LevelString: log.LEVEL_STRING_FATAL,
	}

	runTest(test, to, false, query, 0, 0)
}

func TestEmailLogsEmptyYesEmail(test *testing.T) {
	to := []string{"test1@test.edulinq.org", "test2@test.edulinq.org"}
	query := log.RawLogQuery{
		LevelString: log.LEVEL_STRING_FATAL,
	}

	runTest(test, to, true, query, 1, 0)
}

func runTest(test *testing.T, to []string, sendEmpty bool, query log.RawLogQuery, numMessages int, numRecords int) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	oldValue := log.SetBackgroundLogging(false)
	defer log.SetBackgroundLogging(oldValue)

	log.SetLevels(log.LevelOff, log.LevelTrace)
	defer log.SetLevelFatal()

	// Wait for old logs to get written.
	time.Sleep(10 * time.Millisecond)

	course := db.MustGetTestCourse()

	log.Trace("trace", course)
	log.Debug("debug", course)
	log.Info("info", course)
	log.Warn("warn", course)
	log.Error("error", course)

	task := &tasks.EmailLogsTask{
		BaseTask: &tasks.BaseTask{
			Disable: false,
			When:    []*common.ScheduledTime{},
		},
		To:        to,
		SendEmpty: sendEmpty,
		RawQuery:  query,
	}

	_, err := RunEmailLogsTask(course, task)
	if err != nil {
		test.Fatalf("Failed to run email logs task: '%v'.", err)
	}

	validateMessages(test, numMessages, to, numRecords)
}

func validateMessages(test *testing.T, numMessages int, to []string, numRecords int) {
	messages := email.GetTestMessages()
	email.ClearTestMessages()

	if len(messages) != numMessages {
		test.Fatalf("Did not find the correct number of messages. Expected: %d, Found: %d.", numMessages, len(messages))
	}

	if numMessages == 0 {
		return
	}

	if !reflect.DeepEqual(to, messages[0].To) {
		test.Fatalf("Unexpected message recipients. Expected: [%s], Found: [%s].",
			strings.Join(to, ", "), strings.Join(messages[0].To, ", "))
	}

	lines := strings.Split(strings.TrimSpace(messages[0].Body), "\n")

	expectedLines := numRecords + 2
	if numRecords == 0 {
		expectedLines = 1
	}

	if len(lines) != expectedLines {
		test.Fatalf("Did not find the correct number of lines in the message. Expected: %d, Found: %d.", expectedLines, len(lines))
	}

	countRegex := regexp.MustCompile(`^Found (\d+) log records.*$`)
	groups := countRegex.FindStringSubmatch(lines[0])

	if len(groups) < 2 {
		test.Fatalf("Could not parse the number of matches from the first line of the message: '%s'.", lines[0])
	}

	count, err := strconv.Atoi(groups[1])
	if err != nil {
		test.Fatalf("Failed to parse number of logs ('%s'): '%v'.", groups[1], err)
	}

	if count != numRecords {
		test.Fatalf("Did not find the correct number of log records. Expected: %d, Found: %d.", numRecords, count)
	}
}
