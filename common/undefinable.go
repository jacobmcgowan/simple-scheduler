package common

type Undefinable[T any] struct {
	Value   T
	Defined bool
}
