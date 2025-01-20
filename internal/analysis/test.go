package analysis

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type fakeSimiliartyEngine struct {
	Name string
}

func (this *fakeSimiliartyEngine) GetName() string {
	return this.Name
}

func (this *fakeSimiliartyEngine) IsAvailable() bool {
	return true
}

func (this *fakeSimiliartyEngine) ComputeFileSimilarity(paths [2]string, baseLockKey string) (*model.FileSimilarity, int64, error) {
	similarity := model.FileSimilarity{
		Filename: filepath.Base(paths[0]),
		Tool:     this.Name,
		Score:    float64(len(filepath.Base(paths[0]))) / 100.0,
	}

	return &similarity, 1, nil
}

// Use the fake engines for testing.
// Return a func to reset the engines to their past state.
func UseFakeEnginesForTesting() func() {
	oldEngines := similarityEngines
	resetFunc := func() {
		similarityEngines = oldEngines
	}

	similarityEngines = []core.SimilarityEngine{&fakeSimiliartyEngine{"fake"}}

	return resetFunc
}

func AddTestSubmissions(test *testing.T) {
	// Add the fake student.
	user := db.MustGetServerUser("course-student@test.edulinq.org")
	user.Email = "course-student-alt@test.edulinq.org"
	user.Name = util.StringPointer("course-student-alt")

	err := db.UpsertUser(user)
	if err != nil {
		test.Fatalf("Failed to insert test user: '%v'.", err)
	}

	for _, testSubmission := range testSubmissions {
		assignment := db.MustGetAssignment(testSubmission.Info.CourseID, testSubmission.Info.AssignmentID)

		err = db.SaveSubmission(assignment, testSubmission)
		if err != nil {
			test.Fatalf("Failed to insert test submissions: '%v'.", err)
		}
	}
}

var testSubmissions []*model.GradingResult = []*model.GradingResult{
	&model.GradingResult{
		Info: &model.GradingInfo{
			ID:           "course101::hw0::course-student-alt@test.edulinq.org::1234567890",
			ShortID:      "1234567890",
			CourseID:     "course101",
			AssignmentID: "hw0",
			User:         "course-student-alt@test.edulinq.org",
		},
	},
	&model.GradingResult{
		Info: &model.GradingInfo{
			ID:           "course-languages::bash::course-student-alt@test.edulinq.org::1234567890",
			ShortID:      "1234567890",
			CourseID:     "course-languages",
			AssignmentID: "bash",
			User:         "course-student-alt@test.edulinq.org",
		},
	},
}
