package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestGetLogsLevel(test *testing.T) {
	oldValue := log.SetBackgroundLogging(false)
	defer log.SetBackgroundLogging(oldValue)

	log.SetLevels(log.LevelOff, log.LevelTrace)
	defer log.SetLevelFatal()

	// Wait for old logs to get written.
	time.Sleep(10 * time.Millisecond)

	Clear()
	defer Clear()

	// The logs, records, and levels are all ordered so we can loop over each level.

	log.Trace("trace")
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")

	allRecords := []*log.Record{
		&log.Record{
			log.LevelTrace,
			"trace",
			0, "",
			"", "", "",
			nil,
		},
		&log.Record{
			log.LevelDebug,
			"debug",
			0, "",
			"", "", "",
			nil,
		},
		&log.Record{
			log.LevelInfo,
			"info",
			0, "",
			"", "", "",
			nil,
		},
		&log.Record{
			log.LevelWarn,
			"warn",
			0, "",
			"", "", "",
			nil,
		},
		&log.Record{
			log.LevelError,
			"error",
			0, "",
			"", "", "",
			nil,
		},
	}

	levels := []log.LogLevel{
		log.LevelTrace, log.LevelDebug, log.LevelInfo, log.LevelWarn, log.LevelError, log.LevelFatal,
	}

	for i, level := range levels {
		query := log.ParsedLogQuery{
			Level:        level,
			After:        timestamp.Zero(),
			CourseID:     "",
			AssignmentID: "",
			UserEmail:    "",
		}

		records, err := GetLogRecords(query)
		if err != nil {
			test.Errorf("Level '%s': Failed to get log records: '%v'.", level.String(), err)
			continue
		}

		// Remove the timestamp.
		for _, record := range records {
			record.Timestamp = timestamp.Zero()
		}

		expectedRecords := allRecords[i:len(allRecords)]

		if len(expectedRecords) != len(records) {
			test.Errorf("Level '%s': Got the incorrect number of records. Expected: %d, Actual: %d (%s vs %s).", level.String(),
				len(expectedRecords), len(records),
				util.MustToJSONIndent(expectedRecords), util.MustToJSONIndent(records))
			continue
		}

		if !reflect.DeepEqual(expectedRecords, records) {
			test.Errorf("Level '%s': Unexpected records. Expected: %s, Actual: %s.", level.String(),
				util.MustToJSONIndent(expectedRecords), util.MustToJSONIndent(records))
			continue
		}
	}
}

func (this *DBTests) DBTestGetLogsTime(test *testing.T) {
	Clear()
	defer Clear()

	oldValue := log.SetBackgroundLogging(false)
	defer log.SetBackgroundLogging(oldValue)

	log.SetLevels(log.LevelOff, log.LevelTrace)
	defer log.SetLevelFatal()

	beginning := timestamp.Now()
	time.Sleep(2 * time.Millisecond)

	log.Info("A")
	time.Sleep(2 * time.Millisecond)

	middle := timestamp.Now()
	time.Sleep(2 * time.Millisecond)

	log.Info("B")

	time.Sleep(2 * time.Millisecond)
	end := timestamp.Now()

	allRecords := []*log.Record{
		&log.Record{
			log.LevelInfo,
			"A",
			0, "",
			"", "", "",
			nil,
		},
		&log.Record{
			log.LevelInfo,
			"B",
			0, "",
			"", "", "",
			nil,
		},
	}

	times := []timestamp.Timestamp{
		beginning, middle, end,
	}

	for i, instance := range times {
		query := log.ParsedLogQuery{
			Level:        log.LevelTrace,
			After:        instance,
			CourseID:     "",
			AssignmentID: "",
			UserEmail:    "",
		}

		records, err := GetLogRecords(query)
		if err != nil {
			test.Errorf("Case %d: Failed to get log records: '%v'.", i, err)
			continue
		}

		// Remove the timestamp.
		for _, record := range records {
			record.Timestamp = timestamp.Zero()
		}

		expectedRecords := allRecords[i:len(allRecords)]

		if len(expectedRecords) != len(records) {
			test.Errorf("Case %d: Got the incorrect number of records. Expected: %d, Actual: %d.", i, len(expectedRecords), len(records))
			continue
		}

		if !reflect.DeepEqual(expectedRecords, records) {
			test.Errorf("Case %d: Unexpected records. Expected: %s, Actual: %s.", i,
				util.MustToJSONIndent(expectedRecords), util.MustToJSONIndent(records))
			continue
		}
	}
}

func (this *DBTests) DBTestGetLogsContext(test *testing.T) {
	Clear()
	defer Clear()

	oldValue := log.SetBackgroundLogging(false)
	defer log.SetBackgroundLogging(oldValue)

	log.SetLevels(log.LevelOff, log.LevelTrace)
	defer log.SetLevelFatal()

	log.Info("msg", log.NewCourseAttr("C"))
	log.Info("msg", log.NewCourseAttr("C"), log.NewAssignmentAttr("A"))
	log.Info("msg", log.NewUserAttr("U"))

	courseRecord := &log.Record{log.LevelInfo, "msg", 0, "", "C", "", "", nil}
	assignmentRecord := &log.Record{log.LevelInfo, "msg", 0, "", "C", "A", "", nil}
	userRecord := &log.Record{log.LevelInfo, "msg", 0, "", "", "", "U", nil}

	testCases := []struct {
		courseID        string
		assignmentID    string
		userID          string
		expectedRecords []*log.Record
	}{
		{"C", "", "", []*log.Record{courseRecord, assignmentRecord}},
		{"C", "A", "", []*log.Record{assignmentRecord}},
		{"", "", "U", []*log.Record{userRecord}},

		// Assignment queries must have courses.
		{"", "A", "", []*log.Record{}},

		{"", "", "", []*log.Record{
			courseRecord,
			assignmentRecord,
			userRecord,
		}},
	}

	for i, testCase := range testCases {
		query := log.ParsedLogQuery{
			Level:        log.LevelTrace,
			After:        timestamp.Zero(),
			CourseID:     testCase.courseID,
			AssignmentID: testCase.assignmentID,
			UserEmail:    testCase.userID,
		}

		records, err := GetLogRecords(query)
		if err != nil {
			test.Errorf("Case %d: Failed to get log records: '%v'.", i, err)
			continue
		}

		// Remove the timestamp.
		for _, record := range records {
			record.Timestamp = timestamp.Zero()
		}

		if len(testCase.expectedRecords) != len(records) {
			test.Errorf("Case %d: Got the incorrect number of records. Expected: %d, Actual: %d.", i, len(testCase.expectedRecords), len(records))
			continue
		}

		if !reflect.DeepEqual(testCase.expectedRecords, records) {
			test.Errorf("Case %d: Unexpected records. Expected: %s, Actual: %s.", i,
				util.MustToJSONIndent(testCase.expectedRecords), util.MustToJSONIndent(records))
			continue
		}
	}
}
