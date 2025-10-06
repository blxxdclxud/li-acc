package errs

// Kind represents type of error: System/User/External (err.g. API error)
type Kind int

const (
	// System — infrastructure, dependencies, OS, network, and files errors.
	System Kind = iota

	// User — errors caused by incorrect user actions
	// (err.g., incorrect file format or missing sheet in Excel).
	User

	// External — errors when calling external APIs, SDKs, etc.
	External

	// Validation — input data validation errors (not always systemic).
	Validation
)

// String — helpful method for logs.
func (k Kind) String() string {
	switch k {
	case System:
		return "system"
	case User:
		return "user"
	case External:
		return "external"
	case Validation:
		return "validation"
	default:
		return "unknown"
	}
}
