package main

import (
	"errors"
	"log"
	"net/http"
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

// Returns default value for type
func Zero[T any]() T {
	var zero T
	return zero
}

// Returns last element of array, if array empty, returns error
func Last[T any](arr []T) Result[T] {
	if len(arr) < 1 {
		return ResultErrMessage[T]("Not enough values for unpack")
	}
	return ResultOk(arr[len(arr)-1])
}

func ExceptLast[T any](arr []T) []T {
	if len(arr) < 1 {
		return Zero[[]T]()
	}
	return arr[:len(arr)-1]
}

func UnwrapMessage[T any](msg string, val T, err error) T {
	if err != nil {
		log.Fatalf(msg, err)
	}
	return val
}

// Calls http.Listen and exits the program with log.Fatal
func ListenFatal(addr string, handler http.Handler) {
	log.Fatal(http.ListenAndServe(addr, handler))
}

// Like ListenFatal, but listens multiply addresses simultanously.
// Blocks execution. Can fail if no addresses is passed
func ListenMultiplyFatal(addresses []string) Result[interface{}] {
	if len(addresses) < 1 {
		return ResultErrMessage[interface{}]("No addresses to listen")
	}

	for _, addr := range ExceptLast(addresses) {
		go func(addr string) {
			ListenFatal(addr, nil)
		}(addr)
	}

	ListenFatal(
		Last(addresses).Unwrap(),
		nil)

	return ResultOk[interface{}](nil)
}
