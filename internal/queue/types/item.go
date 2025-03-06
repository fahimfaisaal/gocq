package types

type EnqItem[T any] struct {
	Value    T
	Priority int
	Index    int
}
