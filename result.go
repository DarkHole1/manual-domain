package main

// File for implementing Result[T] from functional programming

import (
	"errors"
	"log"
)

// Result type. Stores value or error
type Result[T any] struct {
	value T
	err   error
}

// Creates ok result from value
func ResultOk[T any](value T) Result[T] {
	return Result[T]{value, nil}
}

// Creates error result from error
func ResultErr[T any](err error) Result[T] {
	return Result[T]{
		Zero[T](),
		err}
}

// Creates error result from error message. Uses errors.New
// internally
func ResultErrMessage[T any](msg string) Result[T] {
	return ResultErr[T](
		errors.New(msg))
}

// If result contains error, writes it in log and exits.
// Othewise returns value.
func (res Result[T]) Unwrap() T {
	if res.err != nil {
		log.Fatal(res.err.Error())
	}
	return res.value
}

// Like Unwrap, but you can provide log message.
// Internally uses Fatalf, giving error in second argument.
func (res Result[T]) UnwrapMessage(msg string) T {
	if res.err != nil {
		log.Fatalf(msg, res.err)
	}
	return res.value
}

// If result is ok returns value, othewise returns argument.
func (res Result[T]) WithDefault(value T) T {
	if res.err != nil {
		return value
	}
	return res.value
}

// Extracts result to pair (value, error)
func (res Result[T]) Extract() (T, error) {
	return res.value, res.err
}
