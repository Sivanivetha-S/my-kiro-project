package models_test

// Tests for models.CalculateAttendancePercentage
//
// Business rule source: requirements.md § 5.7 (Attendance Management),
//                       specifically FR-ATT-03, FR-ATT-04, FR-ATT-05
// Function source:      models/attendance.go
//
// This file contains two kinds of tests:
//
//  1. Example-based (table-driven) tests — verify specific values from the
//     spec including the acceptance criteria example (36/48 → 75.00).
//
//  2. Property-based tests (testing/quick) — generate hundreds of random
//     valid (total, attended) pairs and assert that universal invariants
//     hold. These catch precision bugs, sign errors, and boundary mistakes
//     that no finite set of examples can exercise.

import (
	"testing"
	"testing/quick"

	"studenthub/models"
)

// ---------------------------------------------------------------------------
// 1. Example-based tests
//
// Covers the exact values cited in the spec and coding-standards.md.
// ---------------------------------------------------------------------------

func TestCalculateAttendancePercentage_Examples(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		total      int
		attended   int
		wantPct    float64
	}{
		// Acceptance criteria example (requirements.md § 9, AC-ATT)
		{"AC_ATT_36_of_48", 48, 36, 75.00},

		// tasks.md T-14 acceptance criteria examples
		{"T14_1_of_3",  3,  1,  33.33},
		{"T14_48_of_48", 48, 48, 100.00},

		// Safety net: total = 0 must not panic (design.md § 5.6)
		{"zero_total_zero_attended", 0, 0, 0.0},

		// Rounding: 1/3 = 33.333... → 33.33
		{"one_third", 3, 1, 33.33},

		// Rounding: 2/3 = 66.666... → 66.67
		{"two_thirds", 3, 2, 66.67},

		// Perfect attendance
		{"perfect_attendance_small", 1,   1,   100.00},
		{"perfect_attendance_large", 100, 100, 100.00},

		// Zero attendance
		{"zero_attended",  10, 0, 0.00},
		{"zero_attended_large", 200, 0, 0.00},

		// Threshold boundary: exactly 75% is NOT low attendance
		// (tasks.md T-14: "A student with 75.00% attendance is NOT
		//  in the low-attendance list")
		{"exactly_75pct",    4,  3,  75.00},
		{"exactly_75pct_48", 48, 36, 75.00},

		// One below threshold: 74.99... rounds to 74.99 or similar
		{"just_below_75pct", 100, 74, 74.00},

		// Precision: result must be rounded to exactly 2 decimal places
		{"precision_2dp_a", 7, 3, 42.86}, // 3/7 = 0.42857... → 42.86
		{"precision_2dp_b", 6, 1, 16.67}, // 1/6 = 0.16666... → 16.67
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := models.CalculateAttendancePercentage(tc.total, tc.attended)
			if got != tc.wantPct {
				t.Errorf(
					"CalculateAttendancePercentage(%d, %d) = %.2f, want %.2f",
					tc.total, tc.attended, got, tc.wantPct,
				)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Property-based tests (testing/quick)
//
// Each test states a universal property derived from requirements.md § 5.7.
// testing/quick generates 200 random valid (total, attended) pairs per test.
// ---------------------------------------------------------------------------

// Property 1: The result is always in [0.0, 100.0] for valid inputs.
//
// Business rule: FR-ATT-05 — percentage = (attended / total) * 100.
// Since 0 ≤ attended ≤ total and total ≥ 1, the mathematical result is
// always in [0, 100]. Any value outside that range is a computation error.
//
// Why this matters: The attendance bar component (design.md § 9) renders a
// horizontal bar scaled from 0% to 100%. A value of 110% or −5% would
// break the visual layout without a runtime error — it would silently
// produce a malformed bar.
func TestCalculateAttendancePercentage_Property_AlwaysInValidRange(t *testing.T) {
	property := func(rawTotal, rawAttended uint16) bool {
		total, attended := validAttendancePair(rawTotal, rawAttended)
		pct := models.CalculateAttendancePercentage(total, attended)
		return pct >= 0.0 && pct <= 100.0
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — percentage outside [0, 100]: %v", err)
	}
}

// Property 2: Increasing attended (total held constant) never decreases
// the percentage.
//
// Business rule: requirements.md § 5.7 (FR-ATT-05). The percentage is a
// ratio attended/total. For a fixed total, adding more attended classes
// can only keep the percentage the same or increase it — it cannot decrease.
//
// Why this matters: This property catches a sign flip or an inverted
// division (total/attended instead of attended/total) that would cause all
// computed percentages to be wrong while still producing numbers in [0, 100]
// — and therefore would slip past Property 1.
func TestCalculateAttendancePercentage_Property_MoreAttendedNeverLowersPct(t *testing.T) {
	property := func(rawTotal uint16, rawA, rawB uint16) bool {
		total := int(rawTotal%200) + 1 // total in [1, 200]

		// Two independent attended values, both constrained to [0, total].
		attendedA := int(rawA) % (total + 1)
		attendedB := int(rawB) % (total + 1)

		pctA := models.CalculateAttendancePercentage(total, attendedA)
		pctB := models.CalculateAttendancePercentage(total, attendedB)

		if attendedA >= attendedB {
			return pctA >= pctB
		}
		return pctB >= pctA
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — increasing attended decreased the percentage: %v", err)
	}
}

// Property 3: attended == total always yields 100.00.
//
// Business rule: FR-ATT-05. If a student attended every class
// (attended == total), the percentage must be exactly 100.00, regardless
// of the total class count.
//
// Why this matters: Floating-point division can produce 99.99999... instead
// of 100.0 when the rounding is not applied correctly. This property verifies
// that math.Round(pct*100)/100 correctly produces 100.00 for all totals.
func TestCalculateAttendancePercentage_Property_FullAttendanceIs100(t *testing.T) {
	property := func(rawTotal uint16) bool {
		total := int(rawTotal%1000) + 1 // total in [1, 1000]
		pct := models.CalculateAttendancePercentage(total, total)
		return pct == 100.00
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — full attendance did not yield exactly 100.00: %v", err)
	}
}

// Property 4: attended == 0 always yields 0.00.
//
// Business rule: FR-ATT-04 — attended ≥ 0. When attended is 0, the
// percentage must be exactly 0.00, regardless of total class count.
//
// Why this matters: Same floating-point rounding concern as Property 3,
// applied to the lower boundary.
func TestCalculateAttendancePercentage_Property_ZeroAttendanceIsZero(t *testing.T) {
	property := func(rawTotal uint16) bool {
		total := int(rawTotal%1000) + 1 // total in [1, 1000]
		pct := models.CalculateAttendancePercentage(total, 0)
		return pct == 0.00
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — zero attended did not yield exactly 0.00: %v", err)
	}
}

// Property 5: The result has at most 2 decimal places.
//
// Business rule: FR-ATT-05 — "rounded to two decimal places."
// This property verifies the rounding contract is met for all randomly
// generated inputs, not just the examples in the table above.
//
// Check method: multiply by 100, round to int, divide by 100, compare to
// original — they must be equal (i.e., the original already had ≤ 2 dp).
func TestCalculateAttendancePercentage_Property_MaxTwoDecimalPlaces(t *testing.T) {
	property := func(rawTotal, rawAttended uint16) bool {
		total, attended := validAttendancePair(rawTotal, rawAttended)
		pct := models.CalculateAttendancePercentage(total, attended)

		// Round pct to 2 decimal places independently and compare.
		// If pct already has ≤ 2 dp, re-rounding must not change it.
		reRounded := float64(int(pct*100+0.5)) / 100 // manual round-half-up to 2dp
		// Allow a tiny epsilon for float64 representation tolerance.
		diff := reRounded - pct
		if diff < 0 {
			diff = -diff
		}
		return diff < 1e-9
	}

	if err := quick.Check(property, quickConfig(t)); err != nil {
		t.Errorf("Property violated — result has more than 2 decimal places: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// validAttendancePair converts two arbitrary uint16 values into a valid
// (total, attended) pair satisfying:
//   - total   ∈ [1, 200]   (FR-ATT-03: total_classes ≥ 1)
//   - attended ∈ [0, total] (FR-ATT-04: 0 ≤ attended ≤ total)
//
// uint16 is used as the quick.Check input type instead of int because:
//   - It avoids negative values (which require extra domain mapping).
//   - It keeps the generated range manageable (0–65535) while still
//     covering a wide span of attendance scenarios.
//   - testing/quick generates uint16 values uniformly out of the box.
func validAttendancePair(rawTotal, rawAttended uint16) (total, attended int) {
	total = int(rawTotal%200) + 1     // [1, 200]
	attended = int(rawAttended) % (total + 1) // [0, total]
	return total, attended
}

// quickConfig and hashTestName are defined in marks_test.go.
// Both files share the package models_test so those helpers are visible here.
// (Go allows multiple files in the same test package.)

// Explicit re-declaration guard: quickConfig and hashTestName are declared
// in marks_test.go. Do NOT redeclare them here — that would cause a
// "already declared" compile error. They are shared within package models_test.

// Compile-time verification that we are testing the right package.
// If models.CalculateAttendancePercentage is renamed or removed, this file
// fails to compile immediately rather than silently passing with no tests.
var _ = models.CalculateAttendancePercentage
