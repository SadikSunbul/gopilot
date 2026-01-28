package gopilot

import (
	"errors"
	"testing"
)

func TestExecutionError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewExecutionError("test-function", cause)

	t.Run("Error message format", func(t *testing.T) {
		expected := `gopilot: execution failed for function "test-function": underlying error`
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		if err.Unwrap() != cause {
			t.Error("Unwrap should return the cause")
		}
	})

	t.Run("IsExecutionError returns true", func(t *testing.T) {
		if !IsExecutionError(err) {
			t.Error("IsExecutionError should return true")
		}
	})

	t.Run("IsExecutionError returns false for other errors", func(t *testing.T) {
		if IsExecutionError(cause) {
			t.Error("IsExecutionError should return false for non-ExecutionError")
		}
	})
}

func TestPipelineError(t *testing.T) {
	cause := errors.New("step failed")
	err := NewPipelineError(2, "process-data", cause)

	t.Run("Error message format", func(t *testing.T) {
		expected := "gopilot: pipeline failed at step 2 (process-data): step failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		if err.Unwrap() != cause {
			t.Error("Unwrap should return the cause")
		}
	})

	t.Run("IsPipelineError returns true", func(t *testing.T) {
		if !IsPipelineError(err) {
			t.Error("IsPipelineError should return true")
		}
	})

	t.Run("Step and Function fields", func(t *testing.T) {
		if err.Step != 2 {
			t.Errorf("expected step 2, got %d", err.Step)
		}
		if err.Function != "process-data" {
			t.Errorf("expected function process-data, got %s", err.Function)
		}
	})
}

func TestParameterError(t *testing.T) {
	err := NewParameterError("city", "must be a non-empty string")

	t.Run("Error message format", func(t *testing.T) {
		expected := `gopilot: parameter "city" error: must be a non-empty string`
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("IsParameterError returns true", func(t *testing.T) {
		if !IsParameterError(err) {
			t.Error("IsParameterError should return true")
		}
	})

	t.Run("Fields", func(t *testing.T) {
		if err.Parameter != "city" {
			t.Errorf("expected parameter city, got %s", err.Parameter)
		}
		if err.Reason != "must be a non-empty string" {
			t.Errorf("expected reason 'must be a non-empty string', got %s", err.Reason)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"ErrNilProvider", ErrNilProvider},
		{"ErrNilFunction", ErrNilFunction},
		{"ErrEmptyFunctionName", ErrEmptyFunctionName},
		{"ErrFunctionNotFound", ErrFunctionNotFound},
		{"ErrFunctionExists", ErrFunctionExists},
		{"ErrInvalidParameters", ErrInvalidParameters},
		{"ErrMissingRequiredParam", ErrMissingRequiredParam},
		{"ErrPipelineEmpty", ErrPipelineEmpty},
		{"ErrWorkflowEmpty", ErrWorkflowEmpty},
		{"ErrGenerationFailed", ErrGenerationFailed},
		{"ErrNoResponse", ErrNoResponse},
		{"ErrInvalidResponse", ErrInvalidResponse},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Error("error should not be nil")
			}
			if tc.err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	original := errors.New("original error")
	execErr := NewExecutionError("fn", original)

	t.Run("errors.Is with wrapped error", func(t *testing.T) {
		if !errors.Is(execErr, original) {
			t.Error("errors.Is should find the wrapped error")
		}
	})

	t.Run("errors.As with ExecutionError", func(t *testing.T) {
		var target *ExecutionError
		if !errors.As(execErr, &target) {
			t.Error("errors.As should work with ExecutionError")
		}
		if target.Function != "fn" {
			t.Errorf("expected function 'fn', got %s", target.Function)
		}
	})
}
