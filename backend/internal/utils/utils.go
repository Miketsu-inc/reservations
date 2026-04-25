package utils

import (
	"reflect"
	"time"
)

func StructToMap(data any) map[string]any {
	result := make(map[string]any)

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // unexported field
			continue
		}
		result[field.Name] = val.Field(i).Interface()
	}

	return result
}

func CalculatePercentChange(previous int, current int) int {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return int(float64(current-previous) / float64(previous) * 100)
}

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Repeats the given slice for the given times
func RepeatSlice[T any](s []T, times int) []T {
	if times <= 0 || len(s) == 0 {
		return s
	}

	n := len(s)
	result := make([]T, n*times)
	copy(result, s)

	for copied := n; copied < len(result); copied *= 2 {
		copy(result[copied:], result[:copied])
	}

	return result
}

// Repeats each element for given times
func RepeatEach[T any](s []T, times int) []T {
	if times <= 0 || len(s) == 0 {
		return s
	}

	n := len(s)
	result := make([]T, 0, n*times)
	for _, v := range s {
		for i := 0; i < times; i++ {
			result = append(result, v)
		}
	}

	return result
}
