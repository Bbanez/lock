package utils

type Option[T any] struct {
	Value     T
	Available bool
}

func NewOption[T any](value *T) Option[T] {
	if value == nil {
		return None[T]()
	}
	return Some(*value)
}

func Some[T any](value T) Option[T] {
	return Option[T]{
		Value:     value,
		Available: true,
	}
}

func None[T any]() Option[T] {
	return Option[T]{
		Available: false,
	}
}
