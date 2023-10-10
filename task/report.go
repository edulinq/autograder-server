package task

import (
    "fmt"
    "time"

    "gonum.org/v1/gonum/stat"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/util"
    "github.com/eriq-augustine/autograder/usr"
)

type ReportingSource interface {
    GetUsers() (map[string]*usr.User, error)
    GetAllRecentSubmissionResults(users map[string]*usr.User) (map[string]string, error)
}

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

func GetScoringReport(source ReportingSource) (*ScoringReport, error) {
    questionNames, scores, lastSubmissionTime, err := fetchScores(source);
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

func fetchScores(source ReportingSource) ([]string, map[string][]float64, time.Time, error) {
    users, err := source.GetUsers();
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get users for course: '%w'.", err);
    }

    paths, err := source.GetAllRecentSubmissionResults(users);
    if (err != nil) {
        return nil, nil, time.Time{}, fmt.Errorf("Failed to get submission results: '%w'.", err);
    }

    questionNames := make([]string, 0);
    scores := make(map[string][]float64);
    lastSubmissionTime := time.Time{};

    for email, path := range paths {
        if (users[email].Role != usr.Student) {
            continue;
        }

        if (path == "") {
            continue;
        }

        result := artifact.GradedAssignment{};
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
