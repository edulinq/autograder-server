package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestTaskValidationBase(test *testing.T) {
	courseID := "course101"

	testCases := []struct {
		task                   *UserTaskInfo
		expectedJSON           string
		validateErrorSubstring string
	}{
		// Base

		{
			&UserTaskInfo{
				Type: TaskTypeCourseBackup,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			`{
                "type": "backup",
                "when": {
                    "daily": "3:00",
                    "every": {}
                }
            }`,
			"",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseReport,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
				Options: map[string]any{
					"to": []string{
						"course-admin@test.edulinq.org",
					},
				},
			},
			`{
                "type": "report",
                "when": {
                    "daily": "3:00",
                    "every": {}
                },
                "options": {
                    "to": [
                        "course-admin@test.edulinq.org"
                    ]
                }
            }`,
			"",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseScoringUpload,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			`{
                "type": "scoring-upload",
                "when": {
                    "daily": "3:00",
                    "every": {}
                }
            }`,
			"",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseUpdate,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			`{
                "type": "update",
                "when": {
                    "daily": "3:00",
                    "every": {}
                }
            }`,
			"",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseEmailLogs,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
				Options: map[string]any{
					"query": log.RawLogQuery{
						LevelString: "error",
					},
					"to": []string{
						"course-admin@test.edulinq.org",
					},
				},
			},
			`{
                "type": "email-logs",
                "when": {
                    "daily": "3:00",
                    "every": {}
                },
                "options": {
                    "query": {
                        "level": "error"
                    },
                    "to": [
                        "course-admin@test.edulinq.org"
                    ],
                    "send-empty": false
                }
            }`,
			"",
		},
		{
			&UserTaskInfo{
				Type:     TaskTypeCourseBackup,
				Disabled: true,
			},
			`{
                "type": "backup",
                "disabled": true
            }`,
			"",
		},

		// Errors
		{
			&UserTaskInfo{
				Type: TaskTypeUnknown,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			``,
			"Unknown task type",
		},
		{
			&UserTaskInfo{
				Type: "ZZZ",
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			``,
			"Unknown task type",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseBackup,
			},
			``,
			"Scheduled time to run ('when') is not supplied and the task is not disabled.",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseBackup,
				When: &util.ScheduledTime{
					Daily: "ZZZ",
				},
			},
			``,
			"Failed to validate scheduled time to run",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseReport,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			``,
			"no email recipients are declared",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseReport,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
				Options: map[string]any{
					"to": []any{
						1,
					},
				},
			},
			``,
			"'to' value is not properly formatted",
		},
	}

	for i, testCase := range testCases {
		// Test validation.
		err := testCase.task.Validate()
		if err != nil {
			if testCase.validateErrorSubstring == "" {
				test.Errorf("Case %d: Got an unexpected validation error: '%v'.", i, err)
				continue
			}

			if !strings.Contains(err.Error(), testCase.validateErrorSubstring) {
				test.Errorf("Case %d: Validation error is not as expected. Expected Substring '%s', Actual: '%s'.",
					i, testCase.validateErrorSubstring, err.Error())
				continue
			}

			continue
		}

		formattedExpectedJSON := util.MustFormatJSONObject(testCase.expectedJSON)

		// Test marshalling.
		formattedActualJSON := util.MustFormatJSONObject(util.MustToJSON(testCase.task))
		if formattedExpectedJSON != formattedActualJSON {
			test.Errorf("Case %d: JSON does not match. Expected '%s', Actual: '%s'.\nFormatted Expected: '%s'\nFormatted Actual:   '%s'",
				i, util.MustFormatJSONObjectIndent(formattedExpectedJSON), util.MustFormatJSONObjectIndent(formattedActualJSON),
				formattedExpectedJSON, formattedActualJSON)
			continue
		}

		// Test unmarshaling.
		var newTask UserTaskInfo
		util.MustJSONFromString(formattedActualJSON, &newTask)

		err = newTask.Validate()
		if err != nil {
			test.Errorf("Case %d: Got an unexpected validation error after unmarshal: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(testCase.task, &newTask) {
			test.Errorf("Case %d: Unmarshaled task not as expected. Expected '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.task), util.MustToJSONIndent(newTask))
			continue
		}

		// Test conversion to full tasks.
		fullOldTask, err := testCase.task.ToFullCourseTask(courseID)
		if err != nil {
			test.Errorf("Case %d: Got an unexpected error creating the old full task: '%v'.", i, err)
			continue
		}

		fullNewTask, err := newTask.ToFullCourseTask(courseID)
		if err != nil {
			test.Errorf("Case %d: Got an unexpected error creating the new full task: '%v'.", i, err)
			continue
		}

		if (fullOldTask == nil) && (fullNewTask == nil) {
			continue
		}

		if !reflect.DeepEqual(fullOldTask, fullNewTask) {
			test.Errorf("Case %d: Full tasks not as expected. Old: '%s', New: '%s'.",
				i, util.MustToJSONIndent(fullOldTask), util.MustToJSONIndent(fullNewTask))
			continue
		}

		// Check full task serialization.
		var jsonFullTask FullScheduledTask
		util.MustJSONFromString(util.MustToJSON(fullOldTask), &jsonFullTask)

		err = jsonFullTask.Validate()
		if err != nil {
			test.Errorf("Case %d: Unmarshaled full task had validation error: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(fullOldTask, &jsonFullTask) {
			test.Errorf("Case %d: Unmarshaled full task not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(fullOldTask), util.MustToJSONIndent(jsonFullTask))
			continue
		}
	}
}

func TestTaskToFullCourseTaskBase(test *testing.T) {
	courseID := "course101"

	// Before comparisons are made, timestamps will be adjusted to one of four values:
	// -1 - Before zero.
	//  0 - Zero time (unix epoch).
	//  1 - After zero, but before 2000.
	//  2 - After 2020.
	splitTime := timestamp.MustGuessFromString("2000-01-01T00:00:00Z")

	testCases := []struct {
		userInfo *UserTaskInfo
		// Only non-consistent information will need to be set (the rest is set in the test).
		expectedSystemInfo *SystemTaskInfo
		errorSubstring     string
	}{
		{
			&UserTaskInfo{
				Type: TaskTypeCourseBackup,
				When: &util.ScheduledTime{
					Daily: "3:00",
				},
			},
			&SystemTaskInfo{
				NextRunTime: timestamp.FromMSecs(2),
			},
			"",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseBackup,
				When: &util.ScheduledTime{
					Every: util.DurationSpec{
						Hours: 3,
					},
				},
			},
			&SystemTaskInfo{
				NextRunTime: timestamp.FromMSecs(1),
			},
			"",
		},
	}

	for i, testCase := range testCases {
		err := testCase.userInfo.Validate()
		if err != nil {
			test.Errorf("Case %d: Got an unexpected user info validation error: '%v'.", i, err)
			continue
		}

		task, err := testCase.userInfo.ToFullCourseTask(courseID)
		if err != nil {
			if testCase.errorSubstring == "" {
				test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
				continue
			}

			if !strings.Contains(err.Error(), testCase.errorSubstring) {
				test.Errorf("Case %d: Error is not as expected. Expected Substring '%s', Actual: '%s'.",
					i, testCase.errorSubstring, err.Error())
				continue
			}

			continue
		}

		// Set consistent information.
		testCase.expectedSystemInfo.Source = TaskSourceCourse
		testCase.expectedSystemInfo.CourseID = courseID
		testCase.expectedSystemInfo.Hash = task.SystemTaskInfo.Hash
		testCase.expectedSystemInfo.LastRunTime = timestamp.Zero()

		// Bucket timestamps.
		task.SystemTaskInfo.LastRunTime = bucketTime(splitTime, task.SystemTaskInfo.LastRunTime)
		task.SystemTaskInfo.NextRunTime = bucketTime(splitTime, task.SystemTaskInfo.NextRunTime)

		if !reflect.DeepEqual(testCase.expectedSystemInfo, &task.SystemTaskInfo) {
			test.Errorf("Case %d: System info not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expectedSystemInfo), util.MustToJSONIndent(&task.SystemTaskInfo))
			continue
		}
	}
}

func TestMergeTimesBase(test *testing.T) {
	earlyTime := timestamp.Zero()
	latterTime := timestamp.Now()

	earlyDailyTask := FullScheduledTask{
		UserTaskInfo{
			When: &util.ScheduledTime{
				Daily: "3:00",
			},
		},
		SystemTaskInfo{
			NextRunTime: earlyTime,
		},
	}

	latterDailyTask := FullScheduledTask{
		UserTaskInfo{
			When: &util.ScheduledTime{
				Daily: "3:00",
			},
		},
		SystemTaskInfo{
			NextRunTime: latterTime,
		},
	}

	earlyEveryTask := FullScheduledTask{
		UserTaskInfo{
			When: &util.ScheduledTime{
				Every: util.DurationSpec{
					Hours: 3,
				},
			},
		},
		SystemTaskInfo{
			NextRunTime: earlyTime,
		},
	}

	latterEveryTask := FullScheduledTask{
		UserTaskInfo{
			When: &util.ScheduledTime{
				Every: util.DurationSpec{
					Hours: 3,
				},
			},
		},
		SystemTaskInfo{
			NextRunTime: latterTime,
		},
	}

	// Daily, LHS is early.
	lhs, rhs := earlyDailyTask, latterDailyTask
	lhs.MergeTimes(&rhs)
	if lhs.NextRunTime != earlyTime {
		test.Fatalf("Daily task did not take earlier (lhs) time.")
	}

	// Daily, RHS is early.
	lhs, rhs = latterDailyTask, earlyDailyTask
	lhs.MergeTimes(&rhs)
	if lhs.NextRunTime != earlyTime {
		test.Fatalf("Daily task did not take earlier (rhs) time.")
	}

	// Every, LHS is early.
	lhs, rhs = earlyEveryTask, latterEveryTask
	lhs.MergeTimes(&rhs)
	if lhs.NextRunTime != latterTime {
		test.Fatalf("Every task did not take latter (rhs) time.")
	}

	// Every, RHS is early.
	lhs, rhs = latterEveryTask, earlyEveryTask
	lhs.MergeTimes(&rhs)
	if lhs.NextRunTime != latterTime {
		test.Fatalf("Every task did not take latter (lhs) time.")
	}
}

func bucketTime(splitTime timestamp.Timestamp, actualTime timestamp.Timestamp) timestamp.Timestamp {
	value := int64(0)
	if actualTime < timestamp.Zero() {
		value = -1
	} else if actualTime > splitTime {
		value = 2
	} else if actualTime > timestamp.Zero() {
		value = 1
	}

	return timestamp.FromMSecs(value)
}
