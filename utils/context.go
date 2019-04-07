package utils

// CtxKey is a context key
type CtxKey struct{ int }

var (
	ctxSpinner = CtxKey{1}
)
