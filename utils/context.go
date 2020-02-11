package utils

// CtxKey is a context key
type CtxKey struct{ int }

var (
	// CtxSpinner is the key to a spinner on the context
	CtxSpinner = CtxKey{1}
	// CtxOutWriter can be used to override the output writer of a spinner
	CtxOutWriter = CtxKey{2}
	// CtxErrWriter can be used to override the error writer of a spinner
	CtxErrWriter = CtxKey{3}
)
