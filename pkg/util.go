package frame2

import (
	"context"
	"path"
	"runtime"
)

// If the given context is not nil, return it.
//
// Otherwise, return a default context.
//
// For now, that's a brand new context.Background(), but that might change
func ContextOrDefault(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// Returns the root of the source directory
//
// This assumes that this very file is located at the pkg directory,
// from the source root.  Refactoring may require changes to the code
func SourceRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	ret := path.Join(path.Dir(filename), "..")

	return ret
}
