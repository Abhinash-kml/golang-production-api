package stores

type Store[T any] interface {
	Add() T
	Remove()
	Get() *T
}
