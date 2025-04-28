package analysis

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type fakeSimiliartyEngine struct {
}

func (this *fakeSimiliartyEngine) GetName() string {
	return "fake"
}

func (this *fakeSimiliartyEngine) IsAvailable() bool {
	return true
}

func (this *fakeSimiliartyEngine) ComputeFileSimilarity(paths [2]string, templatePath string, ctx context.Context) (*model.FileSimilarity, error) {
	similarity := model.FileSimilarity{
		Filename: filepath.Base(paths[0]),
		Tool:     this.GetName(),
		Version:  "0.0.1",
		Score:    float64(len(filepath.Base(paths[0]))) / 100.0,
	}

	return &similarity, nil
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
