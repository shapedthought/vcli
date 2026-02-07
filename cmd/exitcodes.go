package cmd

// Exit codes for owlctl commands.
// These codes are used for CI/CD integration to distinguish between
// different outcomes.
const (
	// ExitSuccess indicates the command completed successfully.
	// For diff: no drift detected. For apply: all changes accepted.
	ExitSuccess = 0

	// ExitError indicates a general error occurred.
	// Examples: API failure, invalid spec file, network error.
	ExitError = 1

	// ExitDriftWarning indicates drift was detected at INFO or WARNING severity.
	// Used by diff commands when non-critical changes are found.
	ExitDriftWarning = 3

	// ExitDriftCritical indicates drift was detected at CRITICAL severity.
	// Used by diff commands when security-relevant changes are found.
	ExitDriftCritical = 4

	// ExitPartialApply indicates partial success in batch apply operations.
	// Some resources were applied successfully, but others failed.
	// This exit code is used when applying multiple spec files and some succeed
	// while others fail. It is not used for single-resource operations.
	// Note: VBR API uses all-or-nothing semantics per resource, so field-level
	// partial success is not possible within a single resource.
	ExitPartialApply = 5

	// ExitResourceNotFound indicates the target resource does not exist in VBR.
	// Used when applying to update-only resources (repos, SOBRs, KMS)
	// that must be created via VBR console first.
	ExitResourceNotFound = 6
)

// ApplyOutcome represents the overall outcome of an apply operation
type ApplyOutcome int

const (
	// OutcomeSuccess means all resources applied successfully
	OutcomeSuccess ApplyOutcome = iota
	// OutcomePartial means some resources succeeded, some failed
	OutcomePartial
	// OutcomeAllFailed means all resources failed
	OutcomeAllFailed
	// OutcomeNotFound means resource not found (update-only mode)
	OutcomeNotFound
)

// ExitCodeForOutcome returns the appropriate exit code for an apply outcome
func ExitCodeForOutcome(outcome ApplyOutcome) int {
	switch outcome {
	case OutcomeSuccess:
		return ExitSuccess
	case OutcomePartial:
		return ExitPartialApply
	case OutcomeNotFound:
		return ExitResourceNotFound
	case OutcomeAllFailed:
		return ExitError
	default:
		return ExitError
	}
}

// DetermineApplyOutcome analyzes a slice of ApplyResults and returns the overall outcome.
// An empty results slice returns OutcomeAllFailed as it likely indicates a programming error.
func DetermineApplyOutcome(results []ApplyResult) ApplyOutcome {
	if len(results) == 0 {
		return OutcomeAllFailed
	}

	successCount := 0
	failedCount := 0
	notFoundCount := 0

	for _, r := range results {
		if r.Error == nil {
			successCount++
		} else if r.NotFound {
			notFoundCount++
			failedCount++
		} else {
			failedCount++
		}
	}

	// Single resource not found
	if len(results) == 1 && notFoundCount == 1 {
		return OutcomeNotFound
	}

	// All succeeded
	if failedCount == 0 {
		return OutcomeSuccess
	}

	// All failed
	if successCount == 0 {
		return OutcomeAllFailed
	}

	// Mixed results
	return OutcomePartial
}
