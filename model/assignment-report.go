package model

import (
    "fmt"
    "time"

    "gonum.org/v1/gonum/stat"

    "github.com/eriq-augustine/autograder/util"
)

type ScoringReport struct {
    NumberOfSubmissions int `json:"number-of-submissions"`
    LatestSubmission time.Time `json:"latest-submission"`
    Questions []*ScoringReportQuestionStats `json:"questions"`
}

type ScoringReportQuestionStats struct {
    QuestionName string `json:"question-name"`
    Min float64 `json:"min"`
    Max float64 `json:"max"`
    Median float64 `json:"median"`
    Mean float64 `json:"mean"`
    StdDev float64 `json:"standard-deviation"`
}

func (this *Assignment) GetScoringReport() (*ScoringReport, error) {
    questionNames, scores, lastSubmissionTime, err := this.fetchScores();
    if (err != nil) {
        return nil, err;
    }

    numSubmissions := 0;
    questions := make([]*ScoringReportQuestionStats, 0, len(questionNames));

    for _, questionName := range questionNames {
        min, max := util.MinMax(scores[questionName]);
        mean, stdDev := stat.MeanStdDev(scores[questionName], nil);
        median := util.Median(scores[questionName]);

        stats := &ScoringReportQuestionStats{
            QuestionName: questionName,
            Min: min,
            Max: max,
            Median: median,
            Mean: mean,
            StdDev: stdDev,
        };

        questions = append(questions, stats);
        numSubmissions = len(scores[questionName]);
    }

    report := ScoringReport{
        NumberOfSubmissions: numSubmissions,
        LatestSubmission: lastSubmissionTime,
        Questions: questions,
    };

    return &report, nil;
}

func (this *Assignment) fetchScores() ([]string, map[string][]float64, time.Time, error) {
    users, err := this.Course.GetUsers();
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get users for course: '%w'.", err);
    }

    paths, err := this.GetAllRecentSubmissionResults(users);
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get submission results: '%w'.", err);
    }

    questionNames := make([]string, 0);
    scores := make(map[string][]float64);
    lastSubmissionTime := time.Time{};

    for email, path := range paths {
        if (users[email].Role != Student) {
            continue;
        }

        if (path == "") {
            continue;
        }

        result := GradedAssignment{};
        err = util.JSONFromFile(path, &result);
        if (err != nil) {
            return nil, nil, time.Time{}, fmt.Errorf("Failed to deserialize submission result '%s': '%w'.", path, err);
        }

        if (result.GradingStartTime.After(lastSubmissionTime)) {
            lastSubmissionTime = result.GradingStartTime;
        }

        if (len(questionNames) == 0) {
            for _, question := range result.Questions {
                questionNames = append(questionNames, question.Name);
                scores[question.Name] = make([]float64, 0);
            }
        }

        for _, question := range result.Questions {
            var score float64 = 0.0;
            if (question.MaxPoints != 0.0) {
                score = question.Score / question.MaxPoints;
            }

            scores[question.Name] = append(scores[question.Name], score);
        }
    }

    return questionNames, scores, lastSubmissionTime, nil;
}
