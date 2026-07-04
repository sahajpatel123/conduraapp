package agent

import (
	"context"
)

// Verifier checks the results of executed steps.
type Verifier interface {
	// Verify checks if a step executed correctly.
	Verify(ctx context.Context, step *Step, result *StepResult) (*VerificationResult, error)

	// ShouldRetry determines if a failed step should be retried.
	ShouldRetry(ctx context.Context, result *StepResult, attempt int) bool
}

// VerificationResult is the result of verifying a step.
type VerificationResult struct {
	// Valid indicates whether the verification passed.
	Valid bool

	// Reason is a human-readable reason for failure.
	Reason string

	// ShouldAbort indicates if the entire plan should be aborted.
	ShouldAbort bool
}

// SimpleVerifier is a basic verifier that checks for success.
type SimpleVerifier struct{}

// NewSimpleVerifier creates a new simple verifier.
func NewSimpleVerifier() *SimpleVerifier {
	return &SimpleVerifier{}
}

// Verify checks if a step succeeded.
func (v *SimpleVerifier) Verify(_ context.Context, _ *Step, result *StepResult) (*VerificationResult, error) {
	if result == nil {
		return &VerificationResult{
			Valid:       false,
			Reason:      "no result",
			ShouldAbort: true,
		}, nil
	}

	if result.Error != nil {
		return &VerificationResult{
			Valid:       false,
			Reason:      result.Error.Error(),
			ShouldAbort: false,
		}, result.Error
	}

	return &VerificationResult{
		Valid:       true,
		Reason:      "success",
		ShouldAbort: false,
	}, nil
}

// ShouldRetry determines if a failed step should be retried.
func (v *SimpleVerifier) ShouldRetry(_ context.Context, _ *StepResult, attempt int) bool {
	// Retry up to 3 times
	return attempt < 3
}
