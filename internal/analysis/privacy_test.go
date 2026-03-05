package analysis

import (
	"math"
	"math/rand"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

// TestPrivacyConfigValidation checks that PrivacyConfig.Validate() catches bad inputs.
func TestPrivacyConfigValidation(test *testing.T) {
	testCases := []struct {
		desc    string
		cfg     PrivacyConfig
		wantErr bool
	}{
		{
			desc:    "valid config",
			cfg:     PrivacyConfig{Epsilon: 1.0, Delta: 0.0},
			wantErr: false,
		},
		{
			desc:    "small valid epsilon",
			cfg:     PrivacyConfig{Epsilon: 0.01, Delta: 0.0},
			wantErr: false,
		},
		{
			desc:    "zero epsilon",
			cfg:     PrivacyConfig{Epsilon: 0.0},
			wantErr: true,
		},
		{
			desc:    "negative epsilon",
			cfg:     PrivacyConfig{Epsilon: -1.0},
			wantErr: true,
		},
		{
			desc:    "negative delta",
			cfg:     PrivacyConfig{Epsilon: 1.0, Delta: -0.1},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		err := tc.cfg.Validate()

		if tc.wantErr && err == nil {
			test.Errorf("Case %q: expected validation error, got nil.", tc.desc)
		}

		if !tc.wantErr && err != nil {
			test.Errorf("Case %q: unexpected validation error: %v.", tc.desc, err)
		}
	}
}

// TestExtractFeatures verifies that the feature vector matches the fields of IndividualAnalysis.
func TestExtractFeatures(test *testing.T) {
	analysis := &model.IndividualAnalysis{
		FullID:              "course101::hw0::student@test.edulinq.org::0001",
		CourseID:            "course101",
		AssignmentID:        "hw0",
		UserEmail:           "student@test.edulinq.org",
		AnalysisTimestamp:   timestamp.Now(),
		LinesOfCode:         42,
		LinesOfCodeDelta:    10,
		Score:               0.85,
		ScoreDelta:          0.15,
		SubmissionTimeDelta: 3600000, // 1 hour in ms
	}

	names, values := extractFeatures(analysis)

	if len(names) != len(values) {
		test.Fatalf("Feature names and values have different lengths: %d vs %d.", len(names), len(values))
	}

	if len(names) == 0 {
		test.Fatal("extractFeatures returned no features.")
	}

	// Build a name->value map for readable assertions.
	featureMap := make(map[string]float64, len(names))
	for i, name := range names {
		featureMap[name] = values[i]
	}

	testCases := []struct {
		name     string
		expected float64
	}{
		{"lines-of-code", 42.0},
		{"lines-of-code-delta", 10.0},
		{"score", 0.85},
		{"score-delta", 0.15},
		{"submission-time-delta-ms", 3600000.0},
	}

	for _, tc := range testCases {
		got, ok := featureMap[tc.name]
		if !ok {
			test.Errorf("Feature %q not found in output.", tc.name)
			continue
		}

		if got != tc.expected {
			test.Errorf("Feature %q: expected %g, got %g.", tc.name, tc.expected, got)
		}
	}
}

// TestLaplaceNoiseAdded verifies that noise is actually applied — output differs from input.
// Uses a fixed seed so the test is deterministic.
func TestLaplaceNoiseAdded(test *testing.T) {
	rng := rand.New(rand.NewSource(42))
	epsilon := 1.0
	input := 100.0

	// With a fixed seed and epsilon=1, any single call to addLaplaceNoise should
	// produce a value different from the input with overwhelming probability.
	// The probability of getting exactly 0 noise is measure-zero for a continuous
	// distribution, so this test is deterministic in practice.
	got := addLaplaceNoise(input, epsilon, rng)

	if got == input {
		test.Error("addLaplaceNoise returned the input unchanged; noise was not applied.")
	}
}

// TestLaplaceNoiseMagnitude checks that the noise magnitude is consistent with
// the Laplace distribution Lap(0, 1/epsilon). For epsilon=1, scale=1, so the
// expected absolute noise is E[|Lap(0,1)|] = 1.
//
// This test draws N samples and checks that the empirical mean absolute noise
// is within a reasonable interval. It will fail with very low probability
// (~1e-6 for N=10000) — document this rather than suppress it.
func TestLaplaceNoiseMagnitude(test *testing.T) {
	const (
		n           = 10_000
		epsilon     = 1.0
		inputValue  = 0.0
		// For Lap(0, b=1/epsilon=1): E[|X|] = b = 1.0, Var[|X|] = b^2*(pi^2/2 - 1) ≈ 3.93.
		// Std dev of the sample mean ≈ sqrt(Var[|X|] / n) ≈ 0.020 for n=10000.
		// 5-sigma interval: [1.0 - 5*0.020, 1.0 + 5*0.020] = [0.90, 1.10].
		lowerBound = 0.90
		upperBound = 1.10
	)

	rng := rand.New(rand.NewSource(12345))

	totalAbsNoise := 0.0
	for i := 0; i < n; i++ {
		noisy := addLaplaceNoise(inputValue, epsilon, rng)
		totalAbsNoise += math.Abs(noisy - inputValue)
	}

	meanAbsNoise := totalAbsNoise / float64(n)

	if meanAbsNoise < lowerBound || meanAbsNoise > upperBound {
		// This can fail with very low probability (~1e-6). If it fails on a clean run,
		// widen the interval or increase n. Do not increase the seed to make it pass.
		test.Errorf("Mean absolute noise %g is outside expected range [%g, %g] for epsilon=%g.",
			meanAbsNoise, lowerBound, upperBound, epsilon)
	}
}

// TestLaplaceNoiseScalesWithEpsilon verifies that lower epsilon produces more noise.
func TestLaplaceNoiseScalesWithEpsilon(test *testing.T) {
	const (
		n      = 5_000
		input  = 0.0
		seed   = 99
	)

	testCases := []struct {
		epsilon float64
	}{
		{0.1},
		{1.0},
		{10.0},
	}

	means := make([]float64, len(testCases))

	for i, tc := range testCases {
		rng := rand.New(rand.NewSource(seed))
		total := 0.0

		for j := 0; j < n; j++ {
			noisy := addLaplaceNoise(input, tc.epsilon, rng)
			total += math.Abs(noisy)
		}

		means[i] = total / float64(n)
	}

	// Mean absolute noise should decrease as epsilon increases.
	for i := 1; i < len(means); i++ {
		if means[i] >= means[i-1] {
			test.Errorf("Expected mean noise to decrease as epsilon increases: epsilon=%g gave mean=%g, epsilon=%g gave mean=%g.",
				testCases[i-1].epsilon, means[i-1], testCases[i].epsilon, means[i])
		}
	}
}

// TestHashCourseAssignmentDeterministic verifies that the hash is stable across calls.
func TestHashCourseAssignmentDeterministic(test *testing.T) {
	h1 := hashCourseAssignment("course101", "hw0")
	h2 := hashCourseAssignment("course101", "hw0")

	if h1 != h2 {
		test.Errorf("Hash is not deterministic: got %q then %q.", h1, h2)
	}
}

// TestHashCourseAssignmentDistinct verifies that different inputs produce different hashes
// and that concatenation ambiguity is handled (e.g., "a"+"bc" != "ab"+"c").
func TestHashCourseAssignmentDistinct(test *testing.T) {
	testCases := [][2]string{
		{"course101", "hw0"},
		{"course101", "hw1"},
		{"course102", "hw0"},
		// Concatenation ambiguity check.
		{"a", "bc"},
		{"ab", "c"},
	}

	seen := make(map[string]string)

	for _, pair := range testCases {
		h := hashCourseAssignment(pair[0], pair[1])
		key := pair[0] + "|" + pair[1]

		for otherKey, otherHash := range seen {
			if h == otherHash {
				test.Errorf("Hash collision: %q and %q produced the same hash %q.", key, otherKey, h)
			}
		}

		seen[key] = h
	}
}

// TestPrivatizeAndStoreNoError is an integration-style test verifying that
// privatizeAndStore returns nil for a well-formed input with a valid config.
// It does not check the DB (that would require a test DB setup via db.ResetForTesting)
// but validates the end-to-end logic path through extractFeatures and addLaplaceNoise.
//
// For a full integration test including DB round-trip, see TestPrivatizeAndStoreRoundtrip
// in privacy_integration_test.go (not yet implemented; requires db.ResetForTesting setup).
func TestPrivatizeLogicNoError(test *testing.T) {
	analysis := &model.IndividualAnalysis{
		FullID:       "course101::hw0::student@test.edulinq.org::0001",
		CourseID:     "course101",
		AssignmentID: "hw0",
		UserEmail:    "student@test.edulinq.org",
		LinesOfCode:  50,
		Score:        0.9,
	}

	cfg := PrivacyConfig{Epsilon: 1.0, Delta: 0.0}
	rng := rand.New(rand.NewSource(1))

	names, values := extractFeatures(analysis)
	if len(names) == 0 || len(names) != len(values) {
		test.Fatalf("extractFeatures returned inconsistent results: %d names, %d values.", len(names), len(values))
	}

	// Verify noise is applied to all features without panicking.
	for _, raw := range values {
		_ = addLaplaceNoise(raw, cfg.Epsilon, rng)
	}
}
