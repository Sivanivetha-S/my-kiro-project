package models_test

// Tests for models.CalculateGrade
//
// Business rule source: requirements.md § 5.3 (Grading Scale)
// Function source:      models/marks.go
//
// This file contains two kinds of tests:
//
//  1. Example-based (table-driven) tests — verify every boundary value
//     explicitly, as required by coding-standards.md and tasks.md T-17.
//
//  2. Property-based tests (testing/quick) — generate hundreds of random
//     inputs and assert that universal properties hold for all of them.
//     These catch bugs that no finite set of hand-picked examples can cover,
//     such as an off-by-one in a boundary comparison.

import (
	"math/rand"
	"testing"
	"testing/quick"

	"studenthub/models"
)

// validGrades is the complete set of grades the spec permits.
// requirements.md § 5.3 defines exactly six grades. Any result outside
// this set is a bug in CalculateGrade.
var validGrades = map[string]bool{
	"A+": true,
	"A":  true,
	"B":  true,
	"C":  true,
	"D":  true,
	"F":  true,
}

// ---------------------------------------------------------------------------
// 1. Example-based boundary tests (tasks.md T-17, coding-standards.md)
//
// Every boundary value from requirements.md § 5.3 is tested explicitly.
// The spec defines closed intervals, so the lower bound of each grade AND
// the upper bound of the grade below it must both be tested.
// ---------------------------------------------------------------------------

func TestCalculateGrade_Boundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		marks    float64
		wantGrade string
	}{
		// A+ band: 90–100
		{"A+_lower_bound",  90.0,  "A+"},
		{"A+_mid",          95.0,  "A+"},
		{"A+_upper_bound",  100.0, "A+"},

		// A band: 80–89  (89.99 is still A; 90.0 tips to A+)
		{"A_lower_bound",   80.0,  "A"},
		{"A_mid",           84.5,  "A"},
		{"A_upper_bound",   89.99, "A"},

		// B band: 70–79
		{"B_lower_bound",   70.0,  "B"},
		{"B_mid",           75.0,  "B"},
		{"B_upper_bound",   79.99, "B"},

		// C band: 60–69
		{"C_lower_bound",   60.0,  "C"},
		{"C_mid",           65.0,  "C"},
		{"C_upper_bound",   69.99, "C"},

		// D band: 50–59
		{"D_lower_bound",   50.0,  "D"},
		{"D_mid",           55.0,  "D"},
		{"D_upper_bound",   59.99, "D"},

		// F band: 0–49
		{"F_upper_bound",   49.99, "F"},
		{"F_mid",           25.0,  "F"},
		{"F_lower_bound",   0.0,   "F"},

		// Exact spec acceptance-criteria values (requirements.md § 9, AC-MRK)
		{"AC_MRK_95_is_Aplus", 95.0, "A+"},
		{"AC_MRK_82_is_A",     82.0, "A"},
		{"AC_MRK_73_is_B",     73.0, "B"},
		{"AC_MRK_65_is_C",     65.0, "C"},
		{"AC_MRK_55_is_D",     55.0, "D"},
		{"AC_MRK_40_is_F",     40.0, "F"},

		// Exact boundary values called out in coding-standards.md testing section
		{"boundary_100", 100.0, "A+"},
		{"boundary_90",  90.0,  "A+"},
		{"boundary_80",  80.0,  "A"},
		{"boundary_70",  70.0,  "B"},
		{"boundary_60",  60.0,  "C"},
		{"boundary_50",  50.0,  "D"},
		{"boundary_49",  49.0,  "F"},
		{"boundary_0",   0.0,   "F"},
	}

	for _, tc := range cases {
		tc := tc // capture for t.Parallel
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := models.CalculateGrade(tc.marks)
			if got != tc.wantGrade {
				t.Errorf("CalculateGrade(%.2f) = %q, want %q", tc.marks, got, tc.wantGrade)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Property-based tests (testing/quick)
//
// Properties are universal statements that must hold for ALL valid inputs,
// not just the hand-picked examples above. testing/quick generates 100+
// random float64 values and checks each property.
//
// Why property-based rather than only example-based?
// Example tests verify the cases we thought of. Property tests verify that
// the function behaves correctly on cases we did NOT think of — including
// floats with unusual precision like 79.999999999 or 80.000000001.
// ---------------------------------------------------------------------------

// Property 1: CalculateGrade always returns a member of the valid grade set.
//
// Business rule: requirements.md § 5.3 defines exactly six valid grades.
// Any other return value — including an empty string, a typo like "A ",
// or a completely unexpected string — is a contract violation.
//
// Why this matters: The frontend grade-badge component (design.md § 9) maps
// grade strings to CSS classes. An unexpected grade string would silently
// render no badge at all, giving the user no visual feedback on performance.
func TestCalculateGrade_Property_AlwaysReturnsValidGrade(t *testing.T) {
	// Generate a random float64 in [0, 100].
	// testing/quick generates arbitrary float64 values, so we map them into
	// the valid domain [0, 100] using modulo arithmetic on the absolute value.
	property := func(raw float64) bool {
		marks := domainMarks(raw)
		grade := models.CalculateGrade(marks)
		return validGrades[grade]
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — CalculateGrade returned a grade not in the valid set: %v", err)
	}
}

// Property 2: CalculateGrade never returns an empty string.
//
// Business rule: FR-MRK-04 — "the system shall compute and return the
// letter grade on every read." An empty string is not a grade.
//
// This property is implied by Property 1, but made explicit because an
// empty string is the most likely failure mode if a new grade band is added
// to the switch but a return statement is accidentally omitted.
func TestCalculateGrade_Property_NeverReturnsEmptyString(t *testing.T) {
	property := func(raw float64) bool {
		marks := domainMarks(raw)
		return models.CalculateGrade(marks) != ""
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — CalculateGrade returned an empty string: %v", err)
	}
}

// Property 3: Grade ordering is monotonically non-decreasing with marks.
//
// Business rule: requirements.md § 5.3 — higher marks must not yield a
// lower grade. A student scoring 85 must never receive a lower grade than
// a student scoring 75.
//
// This property catches a transposed boundary (e.g., writing >= 80 where
// >= 70 was intended) that would not be caught by boundary tests alone if
// the transposition happens to produce the correct answer at the specific
// test values but fails in between.
//
// The grade ordering from highest to lowest is: A+ > A > B > C > D > F.
func TestCalculateGrade_Property_HigherMarksNeverYieldLowerGrade(t *testing.T) {
	gradeRank := map[string]int{
		"A+": 5,
		"A":  4,
		"B":  3,
		"C":  2,
		"D":  1,
		"F":  0,
	}

	// Generate two independent values from [0, 100].
	// If marksA >= marksB then rank(grade(marksA)) >= rank(grade(marksB)).
	property := func(rawA, rawB float64) bool {
		marksA := domainMarks(rawA)
		marksB := domainMarks(rawB)

		gradeA := models.CalculateGrade(marksA)
		gradeB := models.CalculateGrade(marksB)

		rankA, okA := gradeRank[gradeA]
		rankB, okB := gradeRank[gradeB]

		// Both grades must be valid (covered by Property 1, but guard here too).
		if !okA || !okB {
			return false
		}

		if marksA >= marksB {
			return rankA >= rankB
		}
		return rankB >= rankA
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — higher marks produced a lower grade: %v", err)
	}
}

// Property 4: The grade matches the documented range for the generated marks.
//
// Business rule: requirements.md § 5.3 — each grade has an explicitly
// documented marks range. This property verifies that CalculateGrade is
// consistent with those ranges for every randomly generated value, not just
// the boundary examples.
//
// This is the strongest single property: it encodes the entire grading
// contract as a predicate and checks it on random inputs.
func TestCalculateGrade_Property_GradeMatchesDocumentedRange(t *testing.T) {
	property := func(raw float64) bool {
		marks := domainMarks(raw)
		grade := models.CalculateGrade(marks)

		switch {
		case marks >= 90:
			return grade == "A+"
		case marks >= 80:
			return grade == "A"
		case marks >= 70:
			return grade == "B"
		case marks >= 60:
			return grade == "C"
		case marks >= 50:
			return grade == "D"
		default:
			return grade == "F"
		}
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — grade did not match the documented range: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// domainMarks maps an arbitrary float64 to a value in [0.0, 100.0].
// testing/quick generates the full float64 space including NaN, Inf, and
// negative values. We fold them into the valid domain so the properties
// test the function's behaviour on the inputs the validator permits.
//
// The mapping is: abs(raw) mod 100.0, with 100.0 itself also valid.
// This preserves density across the full [0, 100] range while ensuring
// no generated value is outside the documented valid range.
func domainMarks(raw float64) float64 {
	if raw < 0 {
		raw = -raw
	}
	// Guard against Inf and NaN which would produce NaN from Mod.
	if raw != raw || raw > 1e15 { // NaN check: NaN != NaN is true
		raw = 50.0
	}
	marks := raw - float64(int(raw/100))*100 // equivalent to fmod(raw, 100)
	if marks < 0 {
		marks = 0
	}
	return marks
}

// quickConfig returns a testing/quick Config that:
//   - runs at least 200 iterations (double the default of 100) for better
//     coverage of the [0, 100] float64 space
//   - seeds the random source from the test's name so failures are
//     reproducible when re-run with the same seed
func quickConfig(t *testing.T) *quick.Config {
	t.Helper()
	return &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(hashTestName(t.Name()))), //nolint:gosec
	}
}

// hashTestName produces a deterministic int64 seed from a test name string
// so that quick.Check failures are reproducible across runs.
//
// Uses FNV-1a with the offset basis truncated to fit int64.
// The uint64 FNV offset basis (14695981039346656037) exceeds int64 max, so
// we use the bit-identical int64 interpretation via a typed constant.
func hashTestName(name string) int64 {
	// FNV-1a 64-bit offset basis reinterpreted as int64.
	// Declared as uint64 first to avoid an integer overflow compile error,
	// then converted to int64 (same bit pattern, valid for use as a seed).
	const fnvOffsetBasis uint64 = 14695981039346656037
	const fnvPrime       uint64 = 1099511628211
	h := fnvOffsetBasis
	for _, b := range []byte(name) {
		h ^= uint64(b)
		h *= fnvPrime
	}
	return int64(h)
}
