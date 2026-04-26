package cmd

import (
	"testing"

	"github.com/ryugen04/sango/internal/doctor"
)

func TestBuildDoctorReportCountsStatuses(t *testing.T) {
	report := buildDoctorReport([]doctor.CheckResult{
		{Name: "git", Status: doctor.StatusPass},
		{Name: "docker", Status: doctor.StatusFail},
		{Name: "sandbox", Status: doctor.StatusWarn},
		{Name: "node", Status: doctor.StatusPass},
	})

	if report.Summary.Passed != 2 {
		t.Fatalf("Passed = %d, want 2", report.Summary.Passed)
	}
	if report.Summary.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", report.Summary.Failed)
	}
	if report.Summary.Warned != 1 {
		t.Fatalf("Warned = %d, want 1", report.Summary.Warned)
	}
	if len(report.Results) != 4 {
		t.Fatalf("Results len = %d, want 4", len(report.Results))
	}
}
