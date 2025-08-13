package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// A key for pairwise analysis.
// Should always be an ordered (lexicographically) pair of full submissions IDs.
type PairwiseKey [2]string

const (
	PAIRWISE_KEY_DELIM    string = "||"
	DEFAULT_INCLUDE_REGEX string = ".+"
)

// Holds options specific to a particular analysis engine.
type OptionsMap map[string]any

type AssignmentAnalysisOptions struct {
	IncludePatterns []string `json:"include-patterns,omitempty,omitzero"`
	ExcludePatterns []string `json:"exclude-patterns,omitempty,omitzero"`

	TemplateFiles   []*util.FileSpec      `json:"template-files,omitempty,omitzero"`
	TemplateFileOps []*util.FileOperation `json:"template-file-ops,omitempty,omitzero"`

	// EngineOptions includes the options for analysis engines.
	// It is a map keyed by an engine name to it's respective engine options.
	// Current supported values can be found in the respective engine's options struct.
	EngineOptions map[string]OptionsMap `json:"engine-options,omitempty,omitzero"`
}

type AnalysisFileInfo struct {
	Filename         string `json:"filename"`
	OriginalFilename string `json:"original-filename,omitempty"`
	LinesOfCode      int    `json:"lines-of-code"`
}

type FileSimilarity struct {
	Filename         string `json:"filename"`
	OriginalFilename string `json:"original-filename,omitempty"`

	Tool    string         `json:"tool"`
	Version string         `json:"version"`
	Options map[string]any `json:"options,omitempty,omitzero"`
	Score   float64        `json:"score"`
}

type AnalysisSummary struct {
	Complete      bool `json:"complete"`
	CompleteCount int  `json:"complete-count"`
	PendingCount  int  `json:"pending-count"`
	FailureCount  int  `json:"failure-count"`
	ErrorCount    int  `json:"error-count"`

	FirstTimestamp timestamp.Timestamp `json:"first-timestamp"`
	LastTimestamp  timestamp.Timestamp `json:"last-timestamp"`
}

type IndividualAnalysis struct {
	Options *AssignmentAnalysisOptions `json:"options,omitempty"`

	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`

	Failure        bool   `json:"failure,omitempty"`
	FailureMessage string `json:"failure-message,omitempty"`

	FullID       string `json:"submission-id"`
	ShortID      string `json:"short-id"`
	CourseID     string `json:"course-id"`
	AssignmentID string `json:"assignment-id"`
	UserEmail    string `json:"user-email"`

	SubmissionStartTime timestamp.Timestamp `json:"submission-start-time,omitempty"`
	Score               float64             `json:"score,omitempty"`

	Files        []AnalysisFileInfo `json:"files,omitempty,omitzero"`
	SkippedFiles []string           `json:"skipped-files,omitempty,omitzero"`
	LinesOfCode  int                `json:"lines-of-code,omitempty"`

	SubmissionTimeDelta int64   `json:"submission-time-delta,omitempty"`
	LinesOfCodeDelta    int     `json:"lines-of-code-delta,omitempty"`
	ScoreDelta          float64 `json:"score-delta,omitempty"`

	LinesOfCodeVelocity float64 `json:"lines-of-code-per-hour,omitempty"`
	ScoreVelocity       float64 `json:"score-per-hour,omitempty"`
}

type IndividualAnalysisSummary struct {
	AnalysisSummary

	AggregateScore util.AggregateValues `json:"aggregate-score"`

	AggregateLinesOfCode        util.AggregateValues            `json:"aggregate-lines-of-code"`
	AggregateLinesOfCodePerFile map[string]util.AggregateValues `json:"aggregate-lines-of-code-per-file"`

	AggregateSubmissionTimeDelta util.AggregateValues `json:"aggregate-submission-time-delta"`
	AggregateLinesOfCodeDelta    util.AggregateValues `json:"aggregate-lines-of-code-delta"`
	AggregateScoreDelta          util.AggregateValues `json:"aggregate-score-delta"`

	AggregateLinesOfCodeVelocity util.AggregateValues `json:"aggregate-lines-of-code-per-hour"`
	AggregateScoreVelocity       util.AggregateValues `json:"aggregate-score-per-hour"`
}

type PairwiseAnalysis struct {
	Options *AssignmentAnalysisOptions `json:"options,omitempty"`

	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`
	SubmissionIDs     PairwiseKey         `json:"submission-ids"`

	Failure        bool   `json:"failure,omitempty"`
	FailureMessage string `json:"failure-message,omitempty"`

	Similarities   map[string][]*FileSimilarity `json:"similarities,omitempty,omitzero"`
	UnmatchedFiles [][2]string                  `json:"unmatched-files,omitempty,omitzero"`
	SkippedFiles   []string                     `json:"skipped-files,omitempty,omitzero"`

	MeanSimilarities    map[string]float64 `json:"mean-similarities,omitempty,omitzero"`
	TotalMeanSimilarity float64            `json:"total-mean-similarity,omitempty"`
}

type PairwiseAnalysisSummary struct {
	AnalysisSummary

	AggregateMeanSimilarities      map[string]util.AggregateValues `json:"aggregate-mean-similarities"`
	AggregateTotalMeanSimilarities util.AggregateValues            `json:"aggregate-total-mean-similarity"`
}

func NewPairwiseKey(fullSubmissionID1 string, fullSubmissionID2 string) PairwiseKey {
	return PairwiseKey([2]string{
		min(fullSubmissionID1, fullSubmissionID2),
		max(fullSubmissionID1, fullSubmissionID2),
	})
}

func (this *AssignmentAnalysisOptions) Validate() error {
	if this == nil {
		return fmt.Errorf("Analysis options cannot be nil.")
	}

	var errs error

	errs = errors.Join(errs, this.validateIncludeExclude())
	errs = errors.Join(errs, this.validateTemplateFiles())

	return errs
}

// Include/Exclude patterns must be valid regular expressions.
// If no include patterns are supplied, DEFAULT_INCLUDE_REGEX is used.
func (this *AssignmentAnalysisOptions) validateIncludeExclude() error {
	var errs error

	if len(this.IncludePatterns) == 0 {
		this.IncludePatterns = append(this.IncludePatterns, DEFAULT_INCLUDE_REGEX)
	}

	for _, pattern := range this.IncludePatterns {
		_, err := regexp.Compile(pattern)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to compile include pattern `%s`: '%w'.", pattern, err))
		}
	}

	for _, pattern := range this.ExcludePatterns {
		_, err := regexp.Compile(pattern)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to compile exclude pattern `%s`: '%w'.", pattern, err))
		}
	}

	return errs
}

// Check the inclusion/exclusion to see if a given relpah is allowed.
func (this *AssignmentAnalysisOptions) MatchRelpath(relpath string) bool {
	match := false
	for _, pattern := range this.IncludePatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(relpath) {
			match = true
			break
		}
	}

	if !match {
		return false
	}

	for _, pattern := range this.ExcludePatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(relpath) {
			return false
		}
	}

	return true
}

func (this *PairwiseKey) String() string {
	return this[0] + PAIRWISE_KEY_DELIM + this[1]
}

func (this PairwiseKey) LogValue() []*log.Attr {
	logAttributes := make([]*log.Attr, 0)

	courseID, assignmentID, email, shortSubmissionID, err := common.SplitFullSubmissionID(this[0])
	if err != nil {
		log.Error("Failed to split submission ID.", err, log.NewAttr("submission", this[0]))

		logAttributes = append(logAttributes, log.NewAttr("submission", this[0]))
	} else {
		logAttributes = append(logAttributes, log.NewCourseAttr(courseID))
		logAttributes = append(logAttributes, log.NewAssignmentAttr(assignmentID))
		logAttributes = append(logAttributes, log.NewUserAttr(email))
		logAttributes = append(logAttributes, log.NewAttr("submission", shortSubmissionID))
	}

	altCourseID, altAssignmentID, altEmail, altShortSubmissionID, err := common.SplitFullSubmissionID(this[1])
	if err != nil {
		log.Error("Failed to split submission ID.", err, log.NewAttr("submission", this[1]))

		logAttributes = append(logAttributes, log.NewAttr("alt-submission", this[1]))
	} else {
		if altCourseID != courseID {
			logAttributes = append(logAttributes, log.NewAttr("alt-course", altCourseID))
		}

		if altAssignmentID != assignmentID {
			logAttributes = append(logAttributes, log.NewAttr("alt-assignment", altAssignmentID))
		}

		if altEmail != email {
			logAttributes = append(logAttributes, log.NewAttr("alt-user", altEmail))
		}

		logAttributes = append(logAttributes, log.NewAttr("alt-submission", altShortSubmissionID))
	}

	return logAttributes
}

// Get the representative course ID for this key.
// Will return an empty string if there is no such course or the ID is malformed.
func (this *PairwiseKey) Course() string {
	if this == nil {
		return ""
	}

	courseID, _, _, _, err := common.SplitFullSubmissionID(this[0])
	if err != nil {
		return ""
	}

	return courseID
}

func (this PairwiseKey) MarshalText() ([]byte, error) {
	keyString := this.String()

	return []byte(keyString), nil
}

func (this *PairwiseKey) UnmarshalText(text []byte) error {
	keyString := string(text)

	keyParts := strings.Split(keyString, PAIRWISE_KEY_DELIM)
	if len(keyParts) != 2 {
		return fmt.Errorf("Invalid PairwiseKey: '%s'.", keyString)
	}

	*this = NewPairwiseKey(keyParts[0], keyParts[1])

	return nil
}

func NewPairwiseAnalysis(pairwiseKey PairwiseKey, assignment *Assignment, similarities map[string][]*FileSimilarity, unmatches [][2]string, skipped []string) *PairwiseAnalysis {
	meanSimilarities := make(map[string]float64, len(similarities))
	totalMeanSimilarity := 0.0

	for relpath, sims := range similarities {
		value := 0.0
		for _, sim := range sims {
			value += sim.Score
		}

		if len(sims) > 0 {
			value /= float64(len(sims))
		}

		meanSimilarities[relpath] = value
		totalMeanSimilarity += value
	}

	if len(similarities) > 0 {
		totalMeanSimilarity /= float64(len(similarities))
	}

	var options *AssignmentAnalysisOptions
	if assignment != nil {
		options = assignment.AssignmentAnalysisOptions
	}

	return &PairwiseAnalysis{
		Options:             options,
		AnalysisTimestamp:   timestamp.Now(),
		SubmissionIDs:       pairwiseKey,
		Similarities:        similarities,
		UnmatchedFiles:      unmatches,
		SkippedFiles:        skipped,
		MeanSimilarities:    meanSimilarities,
		TotalMeanSimilarity: totalMeanSimilarity,
	}
}

func NewFailedPairwiseAnalysis(pairwiseKey PairwiseKey, assignment *Assignment, message string) *PairwiseAnalysis {
	var options *AssignmentAnalysisOptions
	if assignment != nil {
		options = assignment.AssignmentAnalysisOptions
	}

	return &PairwiseAnalysis{
		Options:           options,
		AnalysisTimestamp: timestamp.Now(),
		SubmissionIDs:     pairwiseKey,
		Failure:           true,
		FailureMessage:    message,
	}
}

func NewIndividualAnalysisSummary(results map[string]*IndividualAnalysis, pendingCount int, errorCount int) *IndividualAnalysisSummary {
	if len(results) == 0 {
		return &IndividualAnalysisSummary{
			AnalysisSummary: AnalysisSummary{
				Complete:       (pendingCount == 0),
				CompleteCount:  0,
				PendingCount:   pendingCount,
				FailureCount:   0,
				ErrorCount:     errorCount,
				FirstTimestamp: timestamp.Zero(),
				LastTimestamp:  timestamp.Zero(),
			},
		}
	}

	firstTimestamp := timestamp.Zero()
	lastTimestamp := timestamp.Zero()

	scores := make([]float64, 0, len(results))
	locs := make([]float64, 0, len(results))
	timeDeltas := make([]float64, 0, len(results))
	locDeltas := make([]float64, 0, len(results))
	scoreDeltas := make([]float64, 0, len(results))
	locVelocities := make([]float64, 0, len(results))
	scoreVelocities := make([]float64, 0, len(results))

	locPerFiles := make(map[string][]float64)

	failureCount := 0

	for _, result := range results {
		if result.Failure {
			failureCount++
			continue
		}

		if firstTimestamp.IsZero() || (result.AnalysisTimestamp < firstTimestamp) {
			firstTimestamp = result.AnalysisTimestamp
		}

		if lastTimestamp.IsZero() || (result.AnalysisTimestamp > lastTimestamp) {
			lastTimestamp = result.AnalysisTimestamp
		}

		for _, info := range result.Files {
			locPerFiles[info.Filename] = append(locPerFiles[info.Filename], float64(info.LinesOfCode))
		}

		scores = append(scores, result.Score)
		locs = append(locs, float64(result.LinesOfCode))
		timeDeltas = append(timeDeltas, float64(result.SubmissionTimeDelta))
		locDeltas = append(locDeltas, float64(result.LinesOfCodeDelta))
		scoreDeltas = append(scoreDeltas, result.ScoreDelta)
		locVelocities = append(locVelocities, result.LinesOfCodeVelocity)
		scoreVelocities = append(scoreVelocities, result.ScoreVelocity)
	}

	aggregateLOCPerFile := make(map[string]util.AggregateValues, len(locPerFiles))
	for relpath, locValues := range locPerFiles {
		aggregateLOCPerFile[relpath] = util.ComputeAggregates(locValues)
	}

	return &IndividualAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       (pendingCount == 0),
			CompleteCount:  len(scores),
			PendingCount:   pendingCount,
			FailureCount:   failureCount,
			ErrorCount:     errorCount,
			FirstTimestamp: firstTimestamp,
			LastTimestamp:  lastTimestamp,
		},
		AggregateScore:               util.ComputeAggregates(scores),
		AggregateLinesOfCode:         util.ComputeAggregates(locs),
		AggregateLinesOfCodePerFile:  aggregateLOCPerFile,
		AggregateSubmissionTimeDelta: util.ComputeAggregates(timeDeltas),
		AggregateLinesOfCodeDelta:    util.ComputeAggregates(locDeltas),
		AggregateScoreDelta:          util.ComputeAggregates(scoreDeltas),
		AggregateLinesOfCodeVelocity: util.ComputeAggregates(locVelocities),
		AggregateScoreVelocity:       util.ComputeAggregates(scoreVelocities),
	}
}

func NewPairwiseAnalysisSummary(results map[PairwiseKey]*PairwiseAnalysis, pendingCount int, errorCount int) *PairwiseAnalysisSummary {
	if len(results) == 0 {
		return &PairwiseAnalysisSummary{
			AnalysisSummary: AnalysisSummary{
				Complete:       (pendingCount == 0),
				CompleteCount:  0,
				PendingCount:   pendingCount,
				FailureCount:   0,
				ErrorCount:     errorCount,
				FirstTimestamp: timestamp.Zero(),
				LastTimestamp:  timestamp.Zero(),
			},
		}
	}

	firstTimestamp := timestamp.Zero()
	lastTimestamp := timestamp.Zero()

	meanSims := make(map[string][]float64)
	totalMeanSims := make([]float64, 0, len(results))

	failureCount := 0

	for _, result := range results {
		if result.Failure {
			failureCount++
			continue
		}

		if firstTimestamp.IsZero() || (result.AnalysisTimestamp < firstTimestamp) {
			firstTimestamp = result.AnalysisTimestamp
		}

		if lastTimestamp.IsZero() || (result.AnalysisTimestamp > lastTimestamp) {
			lastTimestamp = result.AnalysisTimestamp
		}

		for relpath, meanSim := range result.MeanSimilarities {
			meanSims[relpath] = append(meanSims[relpath], meanSim)
		}

		totalMeanSims = append(totalMeanSims, result.TotalMeanSimilarity)
	}

	aggregateMeanSimilarities := make(map[string]util.AggregateValues, len(meanSims))
	for relpath, meanSimValues := range meanSims {
		aggregateMeanSimilarities[relpath] = util.ComputeAggregates(meanSimValues)
	}

	return &PairwiseAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       (pendingCount == 0),
			CompleteCount:  len(totalMeanSims),
			PendingCount:   pendingCount,
			FailureCount:   failureCount,
			ErrorCount:     errorCount,
			FirstTimestamp: firstTimestamp,
			LastTimestamp:  lastTimestamp,
		},
		AggregateMeanSimilarities:      aggregateMeanSimilarities,
		AggregateTotalMeanSimilarities: util.ComputeAggregates(totalMeanSims),
	}
}

func (this *IndividualAnalysis) RoundWithPrecision(precision uint) {
	if this == nil {
		return
	}

	this.Score = util.RoundWithPrecision(this.Score, precision)
	this.ScoreDelta = util.RoundWithPrecision(this.ScoreDelta, precision)
	this.LinesOfCodeVelocity = util.RoundWithPrecision(this.LinesOfCodeVelocity, precision)
	this.ScoreVelocity = util.RoundWithPrecision(this.ScoreVelocity, precision)
}

func (this *PairwiseAnalysis) RoundWithPrecision(precision uint) {
	if this == nil {
		return
	}

	for _, sims := range this.Similarities {
		for _, sim := range sims {
			sim.Score = util.RoundWithPrecision(sim.Score, precision)
		}
	}

	this.TotalMeanSimilarity = util.RoundWithPrecision(this.TotalMeanSimilarity, precision)

	for key, value := range this.MeanSimilarities {
		this.MeanSimilarities[key] = util.RoundWithPrecision(value, precision)
	}
}

func (this *IndividualAnalysisSummary) RoundWithPrecision(precision uint) {
	if this == nil {
		return
	}

	this.AggregateScore = this.AggregateScore.RoundWithPrecision(precision)
	this.AggregateLinesOfCode = this.AggregateLinesOfCode.RoundWithPrecision(precision)
	this.AggregateSubmissionTimeDelta = this.AggregateSubmissionTimeDelta.RoundWithPrecision(precision)
	this.AggregateLinesOfCodeDelta = this.AggregateLinesOfCodeDelta.RoundWithPrecision(precision)
	this.AggregateScoreDelta = this.AggregateScoreDelta.RoundWithPrecision(precision)
	this.AggregateLinesOfCodeVelocity = this.AggregateLinesOfCodeVelocity.RoundWithPrecision(precision)
	this.AggregateScoreVelocity = this.AggregateScoreVelocity.RoundWithPrecision(precision)

	for key, sim := range this.AggregateLinesOfCodePerFile {
		this.AggregateLinesOfCodePerFile[key] = sim.RoundWithPrecision(precision)
	}
}

func (this *PairwiseAnalysisSummary) RoundWithPrecision(precision uint) {
	if this == nil {
		return
	}

	this.AggregateTotalMeanSimilarities = this.AggregateTotalMeanSimilarities.RoundWithPrecision(precision)
	for key, sim := range this.AggregateMeanSimilarities {
		this.AggregateMeanSimilarities[key] = sim.RoundWithPrecision(precision)
	}
}

func ComparePairwiseKey(a PairwiseKey, b PairwiseKey) int {
	value := strings.Compare(a[0], b[0])
	if value != 0 {
		return value
	}

	return strings.Compare(a[1], b[1])
}
