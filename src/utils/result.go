package utils

type Result[T any] struct {
	Ok    bool
	Value T
	Error error
}

func Ok[T any](value T) Result[T] {
	return Result[T]{
		Ok:    true,
		Value: value,
	}
}

func Err[T any](err error) Result[T] {
	return Result[T]{
		Ok:    false,
		Error: err,
	}
}

func (r *Result[T]) Unwrap() T {
	if !r.Ok {
		panic(r.Error)
	}
	return r.Value
}
