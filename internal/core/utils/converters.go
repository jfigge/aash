/*
 * Copyright (C) 2024 by Jason Figge
 */

package utils

func SlicePtr[T any](s []T) *[]T {
	return &s
}

func Iff[T interface{}](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func DefaultString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func DefaultInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func DefaultInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
