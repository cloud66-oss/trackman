package utils

// CtxKey is a context key
type CtxKey struct{ int }

var (
	// CtxSpinner is the key to a spinner on the context
	CtxSpinner = CtxKey{1}
)
