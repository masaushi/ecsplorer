package handler

import "context"

// key is a generic type used as a context key for type-safe context values.
type key[T any] struct{}

// TabOption represents an option for handlers that support tabbed interfaces.
type TabOption struct {
	SelectedTabIndex int
}

// contextWithValue adds a value to the context with type safety.
func contextWithValue[T any](ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, key[T]{}, value)
}

// valueFromContext retrieves a value from the context with type safety.
func valueFromContext[T any](ctx context.Context) T {
	val := ctx.Value(key[T]{})
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)
}

// parseTabOption safely extracts the selected tab index from handler options.
func parseTabOption(options []any) int {
	if len(options) == 0 {
		return 0
	}

	if option, ok := options[0].(*TabOption); ok {
		return option.SelectedTabIndex
	}

	return 0
}
