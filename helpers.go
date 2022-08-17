package main

import (
	"log"
	"net/http"
)

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

// Calls http.Listen and exits the program with log.Fatal
func ListenFatal(addr string, handler http.Handler) {
	log.Fatal(http.ListenAndServe(addr, handler))
}

// Like ListenFatal, but listens multiply addresses simultanously.
// Blocks execution. Can fail if no addresses is passed.
// Note: It uses Result, however, it succees with nil. So using Result
// here isn't necessary, but it used for consistency.
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
