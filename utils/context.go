package utils

// CtxKey is a context key
type CtxKey struct{ int }

var (
	CtxSpinner = CtxKey{1}
	CtxLogger  = CtxKey{2}
)
