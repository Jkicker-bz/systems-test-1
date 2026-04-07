package helpers

// Validator accumulates field-level validation errors.
type Validator struct {
	Errors map[string]string
}

func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true when no errors have been recorded.
func (v *Validator) Valid() bool { return len(v.Errors) == 0 }

// Check adds an error for key if condition is false.
func (v *Validator) Check(condition bool, key, message string) {
	if !condition {
		if _, exists := v.Errors[key]; !exists {
			v.Errors[key] = message
		}
	}
}

// Between is a generic helper used by Check calls.
func Between[T int | int64 | float64](value, min, max T) bool {
	return value >= min && value <= max
}
