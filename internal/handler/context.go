package handler

import "context"

type key[T any] struct{}

func contextWithValue[T any](ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, key[T]{}, value)
}

func valueFromContext[T any](ctx context.Context) T {
	val := ctx.Value(key[T]{})
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)
}
