package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
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
				When: &common.ScheduledTime{
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
				When: &common.ScheduledTime{
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
				When: &common.ScheduledTime{
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
				When: &common.ScheduledTime{
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
				Type: TaskTypeEmailLogs,
				When: &common.ScheduledTime{
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
				When: &common.ScheduledTime{
					Daily: "3:00",
				},
			},
			``,
			"Unknown task type",
		},
		{
			&UserTaskInfo{
				Type: "ZZZ",
				When: &common.ScheduledTime{
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
				When: &common.ScheduledTime{
					Daily: "ZZZ",
				},
			},
			``,
			"Failed to validate scheduled time to run",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseReport,
				When: &common.ScheduledTime{
					Daily: "3:00",
				},
			},
			``,
			"no email recipients are declared",
		},
		{
			&UserTaskInfo{
				Type: TaskTypeCourseReport,
				When: &common.ScheduledTime{
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

		if fullOldTask.Hash != fullNewTask.Hash {
			test.Errorf("Case %d: Hashes of old and new full tasks do not match. Old: '%s', New: '%s'.",
				i, util.MustToJSONIndent(fullOldTask), util.MustToJSONIndent(fullNewTask))
			continue
		}
	}
}
